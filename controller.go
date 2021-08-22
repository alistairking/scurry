package scurry

import (
	"context"
	"sync"
	"time"

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
	log    Logger
	cfg    ControllerConfig
	attach *ScAttach

	measQ      chan Measurement
	measCancel context.CancelFunc
	measWg     *sync.WaitGroup

	resQ      chan Measurement
	resCancel context.CancelFunc
	resWg     *sync.WaitGroup
}

func NewController(log zerolog.Logger, cfg ControllerConfig) (*Controller, error) {
	measCtx, measCancel := context.WithCancel(context.Background())
	resCtx, resCancel := context.WithCancel(context.Background())

	attach, err := NewScAttach(log, cfg.ScamperURL)
	if err != nil {
		return nil, err
	}

	c := &Controller{
		log:    initLogger(log, "controller"),
		cfg:    cfg,
		attach: attach,

		measQ:      make(chan Measurement, SEND_Q_LEN),
		measCancel: measCancel,
		measWg:     &sync.WaitGroup{},

		resQ:      make(chan Measurement, RECV_Q_LEN),
		resCancel: resCancel,
		resWg:     &sync.WaitGroup{},
	}

	// start up our measurement execution proxy
	c.measWg.Add(1)
	go c.measurementHandler(measCtx)

	// and our result matching proxy
	c.resWg.Add(1)
	go c.responseHandler(resCtx)

	c.log.Info().
		Interface("config", cfg).
		Msgf("Controller online")

	return c, nil
}

func (c *Controller) MeasurementQueue() chan Measurement {
	return c.measQ
}

func (c *Controller) ResultQueue() chan Measurement {
	return c.resQ
}

func (c *Controller) Drain() {
	c.log.Info().Msgf("Starting drain")
	// the caller should have stopped queueing measurements, so we
	// first wait for our measurement worker to drain
	c.measCancel()
	c.measWg.Wait()

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
	c.log.Info().Msgf("Shutdown complete")
}

func (c *Controller) measurementHandler(ctx context.Context) {
	defer func() {
		close(c.measQ)
		c.measWg.Done()
	}()

	// pull from our measurement queue, convert to a scamper
	// command and hand off to ScAttach for execution
	scQ := c.attach.CommandQueue()
hamster:
	for {
		select {
		case _ = <-c.measQ:
			// TODO: actually convert this to a command
			measStr := "ping 8.8.8.8"
			c.log.Debug().
				Str("measurement", measStr).
				Msgf("Sending measurement to scamper")
			// this might block
			scQ <- measStr

		case <-ctx.Done():
			// canceled, need to drain measQ and then exit
			break hamster
		}
	}

	if len(c.measQ) == 0 {
		return
	}
	c.log.Info().
		Int("queue-length", len(c.measQ)).
		Msgf("Draining measurement queue")
	for _ = range c.measQ {
		// TODO: actually convert this to a command
		measStr := "ping 8.8.8.8"
		c.log.Debug().
			Str("measurement", measStr).
			Msgf("Sending measurement to scamper")
		// this might block
		scQ <- measStr
	}
	c.log.Info().
		Msgf("Measurement queue drained")
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
			c.log.Debug().
				Str("result", resStr).
				Msgf("Received result from scamper")
			// TODO: parse it, match against queued
			// measurement, and then push into c.resQ

		case errStr := <-errQ:
			c.log.Error().
				Str("error", errStr).
				Msgf("Received error from scamper")
			// TODO: handle these in-band

		case <-ctx.Done():
			// canceled, need to drain both queues
			break hamster
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), SHUTDOWN_LINGER)
	defer cancel()
	// TODO: change this so that we stop early once all
	// outstanding measurements are back
	rem := 1 // XXX
	c.log.Info().
		Int("outstanding", rem).
		Dur("linger", SHUTDOWN_LINGER).
		Msgf("Draining scamper response queues")
	for {
		select {
		case resStr := <-resultQ:
			c.log.Debug().
				Str("result", resStr).
				Msgf("Received result from scamper")
			// TODO: parse it, match against queued
			// measurement, and then push into c.resQ
			if rem == 0 {
				// done
				c.log.Info().
					Msgf("Received responses from scamper")
			}

		case errStr := <-errQ:
			c.log.Error().
				Str("error", errStr).
				Msgf("Received error from scamper")
			// TODO: handle these in-band

		case <-ctx.Done():
			c.log.Error().
				Int("outstanding", rem).
				Msgf("Giving up waiting for responses from scamper")
			return
		}
	}
}
