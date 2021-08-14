package scurry

import (
	"net"
	"strings"

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

	sConn net.Conn

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

	// connect and attach to scamper daemon
	err := c.initScamper()
	if err != nil {
		return nil, err
	}

	c.log.Info().
		Interface("config", cfg).
		Msgf("Controller online")

	return c, nil
}

func (c *Controller) initScamper() error {
	// TODO: better unix socket detection
	var conn net.Conn
	var err error
	if strings.Contains(":", c.cfg.ScamperURL) {
		conn, err = net.Dial("tcp", c.cfg.ScamperURL)
	} else {
		conn, err = net.Dial("unix", c.cfg.ScamperURL)
	}
	if err != nil {
		return err
	}
	c.sConn = conn

	// create buffered writer/reader

	// send our attach command now
	return c.sendCmd("attach \"1\" \"json\"")
}

func (c *Controller) sendCmd(cmd string) error {
	c.log.Debug().
		Str("command", cmd).
		Msgf("Sending command to scamper")
	return nil
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

	c.sConn.Close()
}
