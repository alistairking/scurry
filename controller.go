package scurry

import (
	"github.com/rs/zerolog"
)

const (
	// TODO: make this configurable
	SEND_Q_LEN = 100
	RECV_Q_LEN = 100
)

type ControllerConfig struct {
	ScamperURL string
}

// Simple scamper control socket client
type Controller struct {
	log zerolog.Logger
	cfg ControllerConfig

	measQ chan Measurement
	resQ  chan Measurement
}

// TODO: move this to a common location
func initLogger(log zerolog.Logger) zerolog.Logger {
	return log.With().
		Str("package", "scurry").
		Str("module", "controller").
		Logger()
}

func NewController(log zerolog.Logger, cfg ControllerConfig) (*Controller, error) {
	c := &Controller{
		log: initLogger(log),
		cfg: cfg,

		measQ: make(chan Measurement, SEND_Q_LEN),
		resQ:  make(chan Measurement, RECV_Q_LEN),
	}

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
	c.log.Debug().Msgf("Starting drain")
	close(c.measQ)
	// TODO: tell result guy to close the channel once all outstanding measurements come back
	close(c.resQ) // XXXX
}

func (c *Controller) Close() {
	if c == nil {
		return
	}
	c.log.Info().Msgf("Shutting down")
}
