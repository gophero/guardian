package log

import (
	"log"

	"github.com/rs/zerolog"
)

type logWriter struct {
	level  zerolog.Level
	logger zerolog.Logger
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	if n > 0 && p[n-1] == '\n' {
		// Trim CR added by stdlog.
		p = p[0 : n-1]
	}
	w.logger.WithLevel(w.level).CallerSkipFrame(3).Msg(string(p))
	return
}

// NewStdLog creates new [log.Logger] which writes to provided logger at provided level.
func NewStdLog(logger zerolog.Logger, level zerolog.Level) *log.Logger {
	return log.New(&logWriter{logger: logger, level: level}, "", 0)
}
