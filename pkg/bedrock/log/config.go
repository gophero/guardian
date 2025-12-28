package log

import (
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog"
)

type Config struct {
	Level  string `help:"Log level." name:"level" env:"LEVEL" enum:"trace,debug,info,warn,error,fatal,panic,disabled" default:"info"`
	Out    string `help:"Where to write logs." name:"out" env:"OUT" enum:"stderr,stdout" default:"stderr"`
	Pretty bool   `help:"Print colorized and human-friendly log output. This is not performant and should only be used in development." name:"pretty" env:"PRETTY" default:"false"`
}

func (c Config) level() (zerolog.Level, error) {
	lvl, err := zerolog.ParseLevel(c.Level)
	if err != nil {
		return zerolog.NoLevel, fmt.Errorf("logger: zerolog parse level: %w", err)
	}

	return lvl, nil
}

func (c Config) writer() (io.Writer, error) {
	w, err := parseOut(c.Out)
	if err != nil {
		return nil, err
	}

	if c.Pretty {
		w = zerolog.NewConsoleWriter(func(cw *zerolog.ConsoleWriter) {
			cw.Out = w
		})
	}

	return w, nil
}

func parseOut(out string) (io.Writer, error) {
	switch out {
	case "stdout":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	default:
		return nil, fmt.Errorf("logger: `%s` is not a valid output option", out)
	}
}
