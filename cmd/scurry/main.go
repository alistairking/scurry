package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	_ "github.com/alistairking/scurry"
	"github.com/rs/zerolog"
)

type PingCmd struct {
	// TODO
}

type ScurryCLI struct {
	// measurement commands
	Ping PingCmd `cmd help:"Ping measurements"`
	// TODO traceroute

	// global measurement config
	Target []string `help:"IP to execute measurements towards"`

	// scamper connection info
	ScamperURL string `help:"URL to connect to scamper on (host:port or unix domain socket)"`

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

func main() {
	var cliCfg ScurryCLI
	k := kong.Parse(&cliCfg)
	k.Validate()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log, err := initLogger(cliCfg)
	k.FatalIfErrorf(err)

	handleSignals(ctx, log, cancel)

	// TODO: stuff here
	log.Info().
		Interface("cfg", cliCfg).
		Msgf("Scurrying!")

	time.Sleep(time.Second) // wait for logger to drain
}
