package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"runtime"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestStdLog(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	logger := NewStdLog(zerolog.New(buf).With().Caller().Logger(), zerolog.ErrorLevel)
	_, file, line, _ := runtime.Caller(0)
	caller := fmt.Sprintf("%s:%d", file, line+2)
	logger.Printf("std log via %s\n", "zerolog")

	j := make(map[string]any)
	err := json.NewDecoder(buf).Decode(&j)

	require.NoError(t, err, "log should a proper json object")
	require.Equal(t, "error", j["level"], "logged level didn't match expected level")
	require.Equal(t, "std log via zerolog", j["message"], "logged message didn't match expected message")
	require.Equal(t, caller, j["caller"], "logged caller didn't match expected caller")
}
