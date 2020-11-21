package log

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// Level is a log level type used for mapping between string
// representations, and zerolog log levels.
type Level string

func (l Level) ZerologLevel() zerolog.Level {
	switch l {
	case LevelInfo:
		return zerolog.InfoLevel
	case LevelDebug:
		return zerolog.DebugLevel
	case LevelWarn:
		return zerolog.WarnLevel
	case LevelError:
		return zerolog.ErrorLevel
	}

	// default to the lowest level
	return zerolog.DebugLevel
}

const (
	LevelInfo  Level = "info"
	LevelDebug Level = "debug"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// New creates, configures and returns a new zerolog.Logger.
func New(pretty bool, module string, level Level) zerolog.Logger {
	var output io.Writer = os.Stdout
	if pretty {
		output = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	}

	return zerolog.New(output).With().
		Str("mod", module).
		Timestamp().
		Logger().Level(zerolog.Level(level.ZerologLevel()))
}

// ParseLogLevel maps a log level string to a constant which may be used
// to create a new logger.
func ParseLogLevel(input string) (Level, error) {
	input = strings.ToLower(input)
	switch input {
	case string(LevelInfo):
		return LevelInfo, nil
	case string(LevelDebug):
		return LevelDebug, nil
	case string(LevelWarn):
		return LevelWarn, nil
	case string(LevelError):
		return LevelError, nil
	default:
		return LevelInfo, fmt.Errorf("unrecognized log level '%s'", input)
	}
}
