package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/alistairking/scurry"
	"github.com/alistairking/scurry/measurement"
	"github.com/rs/zerolog"
)

type PingCmd struct {
	// TODO
}

type TraceCmd struct {
	// TODO
}

type ScurryCLI struct {
	// measurement commands
	Ping  measurement.Ping  `cmd help:"Ping measurements"`
	Trace measurement.Trace `cmd help:"Traceroute measurements"`

	// global measurement config
	Target []string `required short:"t" help:"IP to execute measurements towards"`
	// TODO: TargetFile
	//
	// scamper connection info
	ScamperURL string `required short:"s" help:"URL to connect to scamper on (host:port or unix domain socket)"`

	// misc flags
	LogLevel string `help:"Log level" default:"info"`
}

func initLogger(cfg ScurryCLI) (zerolog.Logger, error) {
	zl := zerolog.Logger{}
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		return zl, err
	}

	output := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}

	l := zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Logger()
	return l, nil
}

func handleSignals(ctx context.Context, log zerolog.Logger, cancel context.CancelFunc) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-sigCh:
				log.Info().Msgf("Signal recevied, triggering shutdown")
				cancel()
				return
			}
		}
	}()
}

// TODO: turn this inside out so that we let kong call Ping.Run which
// populates the task and then calls a common function to actually do
// the work
func initTask(cmd string, cfg ScurryCLI) (measurement.Task, error) {
	task := measurement.Task{}
	// build a base task that we'll reuse
	typeStr := cmd
	tType, err := measurement.TypeString(typeStr)
	if err != nil {
		return task, err
	}
	task.Type = tType

	switch tType {
	case measurement.TYPE_PING:
		task.Options.Ping = cfg.Ping

	case measurement.TYPE_TRACE:
		task.Options.Trace = cfg.Trace
	}

	return task, nil
}

// TODO: move this stuff into the scurry package? Some kind of
// QueueTargets(ctx, meas, targets) method that does this work. How to
// keep it async?
func queueTasks(ctx context.Context, log zerolog.Logger, wg *sync.WaitGroup,
	ctrl *scurry.Controller, task measurement.Task, cfg ScurryCLI) {
	defer wg.Done()

	mCh := ctrl.TaskQueue()
	for _, target := range cfg.Target {
		log.Debug().
			Str("target", target).
			Msgf("Queueing task")
		task.Target = target
		mCh <- task
	}

	log.Debug().Msgf("Finished queueing tasks")
}

func recvResults(ctx context.Context, log zerolog.Logger, wg *sync.WaitGroup,
	ctrl *scurry.Controller) {
	log.Debug().Msgf("Result receiver online")
	defer wg.Done()

	cnt := uint64(0)
	q := ctrl.ResultQueue()
	eoq := measurement.Task{}
	for {
		select {
		case result := <-q:
			if result == eoq {
				log.Info().
					Uint64("total", cnt).
					Msgf("Finished receiving results")
				return
			}
			cnt++
			j, err := result.AsJson()
			if err != nil {
				log.Error().
					Err(err).
					Msgf("Failed to convert measurement to JSON")
				continue
			}
			fmt.Println(j)
		case <-ctx.Done():
			// canceled, just give up
			return
		}
	}
}

func main() {
	// Parse command line args
	var cliCfg ScurryCLI
	k := kong.Parse(&cliCfg)
	k.Validate()

	// Set up context, logger, and signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log, err := initLogger(cliCfg)
	k.FatalIfErrorf(err)
	handleSignals(ctx, log, cancel)

	// Create a reusable task object.
	// We'll just modify the `Target` field.
	task, err := initTask(k.Command(), cliCfg)
	k.FatalIfErrorf(err)

	// Create the scurry Controller
	ctrl, err := scurry.NewController(log,
		scurry.ControllerConfig{
			ScamperURL: cliCfg.ScamperURL,
		},
	)
	k.FatalIfErrorf(err)
	defer ctrl.Close()

	// Ready to go!
	log.Info().
		Interface("cfg", cliCfg).
		Msgf("Scurrying!")

	// Kick off a goroutine to feed our tasks to scurry
	//
	// We do this asynchronously in case we have more targets than
	// scamper can accept at once.
	qWg := &sync.WaitGroup{}
	qWg.Add(1)
	go queueTasks(ctx, log, qWg, ctrl, task, cliCfg)

	// And another to retrieve the responses
	resWg := &sync.WaitGroup{}
	resWg.Add(1)
	go recvResults(ctx, log, resWg, ctrl)

	// Wait until we have queued all our tasks
	qWg.Wait()

	// Tell the controller that we're done queueing things. This
	// will block until all of the tasks we queued have been
	// handed off to scamper.
	ctrl.Drain()

	// Wait until we've received all the results (the Controller
	// will signal this by closing the result channel).
	resWg.Wait()

	// Wait a moment for the logger to drain any remaining messages
	time.Sleep(time.Second)

	// NB: ctrl.Close() was defer'd above
}
