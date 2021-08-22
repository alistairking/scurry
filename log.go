package scurry

import (
	"github.com/rs/zerolog"
)

type Logger = zerolog.Logger

func initLogger(log Logger, module string) Logger {
	return log.With().
		Str("package", "scurry").
		Str("module", module).
		Logger()
}
