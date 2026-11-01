package witcli

import (
	"os"

	"github.com/kjanat/go-modules/wit/logging"
)

// Logger returns a [logging.Logger] that outputs to stdout.
func Logger(verbose, debug bool) logging.Logger {
	level := logging.LevelWarn
	if debug {
		level = logging.LevelDebug
	} else if verbose {
		level = logging.LevelInfo
	}
	return logging.NewLogger(os.Stderr, level)
}
