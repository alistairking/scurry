package scurry

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alistairking/scurry/measurement"
	"github.com/rs/zerolog"
)

const (
	// TODO: make this configurable
	SEND_Q_LEN      = 100
	RECV_Q_LEN      = 100
	SHUTDOWN_LINGER = time.Second * 60
)

type ControllerConfig struct {
	ScamperURL string
}

// Simple scamper control socket client
type Controller struct {
	log         Logger
	cfg         ControllerConfig
	attach      *ScAttach
	outstanding map[uint64]measurement.Task
	nextId      uint64
	errCmds     uint64
	mu          *sync.RWMutex

	taskQ      chan measurement.Task
	taskCancel context.CancelFunc
	taskWg     *sync.WaitGroup

	resQ      chan measurement.Task
	resCancel context.CancelFunc
	resWg     *sync.WaitGroup
}

func NewController(log zerolog.Logger, cfg ControllerConfig) (*Controller, error) {
	taskCtx, taskCancel := context.WithCancel(context.Background())
	resCtx, resCancel := context.WithCancel(context.Background())

	attach, err := NewScAttach(log, cfg.ScamperURL)
	if err != nil {
		return nil, err
	}

	c := &Controller{
		log:         initLogger(log, "controller"),
		cfg:         cfg,
		attach:      attach,
		outstanding: map[uint64]measurement.Task{},
		nextId:      1,
		mu:          &sync.RWMutex{},

		taskQ:      make(chan measurement.Task, SEND_Q_LEN),
		taskCancel: taskCancel,
		taskWg:     &sync.WaitGroup{},

		resQ:      make(chan measurement.Task, RECV_Q_LEN),
		resCancel: resCancel,
		resWg:     &sync.WaitGroup{},
	}

	// start up our task execution proxy
	c.taskWg.Add(1)
	go c.taskHandler(taskCtx)

	// and our result matching proxy
	c.resWg.Add(1)
	go c.responseHandler(resCtx)

	c.log.Debug().
		Interface("config", cfg).
		Msgf("Controller online")

	return c, nil
}

func (c *Controller) TaskQueue() chan measurement.Task {
	return c.taskQ
}

func (c *Controller) ResultQueue() chan measurement.Task {
	return c.resQ
}

func (c *Controller) Drain() {
	// the caller should have stopped queueing tasks, so we
	// first wait for our task worker to drain
	c.taskCancel()
	c.taskWg.Wait()

	// now signal to the result worker that it should shut down
	// once all outstanding results are back
	c.resCancel()
}

func (c *Controller) Close() {
	if c == nil {
		return
	}
	// wait for result drain to complete (it should be)
	c.resWg.Wait()
	// close our scamper handler
	c.attach.Close()
	c.log.Debug().Msgf("Shutdown complete")
}

func (c *Controller) sendTask(task measurement.Task) {
	c.mu.Lock()
	// TODO: more complex IDs?
	task.UserId = c.nextId
	c.nextId++
	c.outstanding[task.UserId] = task
	c.mu.Unlock()
	taskCmd := task.AsCommand()
	c.log.Debug().
		Interface("task", task).
		Str("command", taskCmd).
		Msgf("Sending command to scamper")
	// this might block
	c.attach.CommandQueue() <- taskCmd
}

func (c *Controller) taskHandler(ctx context.Context) {
	defer func() {
		close(c.taskQ)
		c.taskWg.Done()
	}()

	// pull from our task queue, convert to a scamper
	// command and hand off to ScAttach for execution
hamster:
	for {
		select {
		case task := <-c.taskQ:
			c.sendTask(task)

		case <-ctx.Done():
			// canceled, need to drain taskQ and then exit
			break hamster
		}
	}

	if len(c.taskQ) == 0 {
		return
	}
	c.log.Debug().
		Int("queue-length", len(c.taskQ)).
		Msgf("Draining task queue")
	for len(c.taskQ) > 0 {
		c.sendTask(<-c.taskQ)
	}
	c.log.Debug().
		Msgf("Task queue drained")
}

func (c *Controller) handleResult(resStr string) {
	scRes, err := measurement.NewScResultFromJson(resStr)
	if err != nil {
		// TODO: handle these in-band
		c.log.Error().
			Err(err).
			Msgf("Failed to parse result from scamper")
		return
	}

	// discard the initial cycle-start we get after we attach
	if scRes.Type == "cycle-start" {
		c.log.Debug().
			Str("result", resStr).
			Msgf("Discarding cycle-start")
		return
	}

	c.mu.Lock()
	task, exists := c.outstanding[scRes.UserID]
	delete(c.outstanding, scRes.UserID)
	c.mu.Unlock()

	if !exists {
		c.log.Error().
			Interface("sc-result", scRes).
			Uint64("userid", scRes.UserID).
			Msgf("Couldn't find task for scamper result")
		return
	}

	task.Result = scRes
	c.resQ <- task
}

func (c *Controller) handleError(errStr string) {
	c.log.Error().
		Str("error", errStr).
		Msgf("Received error from scamper")

	// if we get a 'command not accepted' error, we don't know
	// which command exactly failed, but we can increment a
	// counter so that when we shut down we don't bother to wait
	// for this result
	atomic.AddUint64(&c.errCmds, 1)

	// TODO: handle these in-band
}

func (c *Controller) Outstanding() int {
	c.mu.RLock()
	rem := len(c.outstanding)
	c.mu.RUnlock()
	errs := atomic.LoadUint64(&c.errCmds)
	if rem >= int(errs) {
		return rem - int(errs)
	}
	// this is odd
	c.log.Warn().
		Int("outstanding", rem).
		Uint64("errors", errs).
		Msgf("More errors than outstanding tasks")
	return rem
}

func (c *Controller) responseHandler(ctx context.Context) {
	defer func() {
		close(c.resQ)
		c.resWg.Done()
	}()

	// service both the result and error queues from ScAttach
	resultQ := c.attach.ResultQueue()
	errQ := c.attach.ErrorQueue()
hamster:
	for {
		select {
		case resStr := <-resultQ:
			c.handleResult(resStr)

		case errStr := <-errQ:
			c.handleError(errStr)

		case <-ctx.Done():
			// canceled, need to drain both queues
			break hamster
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), SHUTDOWN_LINGER)
	defer cancel()
	rem := c.Outstanding()
	c.log.Info().
		Int("outstanding", rem).
		Dur("linger", SHUTDOWN_LINGER).
		Msgf("Waiting for remaining tasks to complete")
drain:
	for {
		select {
		case resStr := <-resultQ:
			c.handleResult(resStr)
			if c.Outstanding() == 0 {
				// done
				break drain
			}

		case errStr := <-errQ:
			c.handleError(errStr)

		case <-ctx.Done():
			c.log.Error().
				Msgf("Giving up waiting for results from scamper")
			break drain
		}
	}

	// dump any tasks still outstanding back to the user
	// these could be errors, or things that we gave up waiting for
	c.log.Debug().
		Int("abandoned", len(c.outstanding)).
		Msgf("Received all results from scamper")
	for _, task := range c.outstanding {
		c.resQ <- task
	}
}
