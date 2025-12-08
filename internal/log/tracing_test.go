package log

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/embedded"
)

func TestTracingHook(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	logger := zerolog.New(buf).Level(zerolog.InfoLevel).Hook(tracingHook{})

	var traceID trace.TraceID
	var spanID trace.SpanID
	rand.Read(traceID[:])
	rand.Read(spanID[:])

	t.Run("no-context", func(t *testing.T) {
		logger.Info().Msg("test")

		require.JSONEq(t, `{"level":"info","message":"test"}`, buf.String())
	})
	buf.Reset()

	t.Run("with-no-span", func(t *testing.T) {
		logger.Info().Ctx(context.Background()).Msg("test")

		require.JSONEq(t, `{"level":"info","message":"test"}`, buf.String())
	})
	buf.Reset()

	t.Run("with-non-recording-span", func(t *testing.T) {
		ctx := trace.ContextWithSpan(context.Background(), &testSpan{isRecording: false, sc: trace.NewSpanContext(trace.SpanContextConfig{
			TraceID: traceID,
			SpanID:  spanID,
		})})

		logger.Info().Ctx(ctx).Msg("test")

		require.JSONEq(t, `{"level":"info","message":"test"}`, buf.String())
	})
	buf.Reset()

	t.Run("with-invalid-span-context", func(t *testing.T) {
		ctx := trace.ContextWithSpan(context.Background(), &testSpan{isRecording: true, sc: trace.NewSpanContext(trace.SpanContextConfig{})})

		logger.Info().Ctx(ctx).Msg("test")

		require.JSONEq(t, `{"level":"info","message":"test"}`, buf.String())
	})
	buf.Reset()

	t.Run("with-valid-span-context", func(t *testing.T) {
		ctx := trace.ContextWithSpan(context.Background(), &testSpan{isRecording: true, sc: trace.NewSpanContext(trace.SpanContextConfig{
			TraceID: traceID,
			SpanID:  spanID,
		})})

		logger.Info().Ctx(ctx).Msg("test")

		require.JSONEq(t, fmt.Sprintf(`{"level":"info","message":"test","trace_id":"%s","span_id":"%s"}`, traceID.String(), spanID.String()), buf.String())
	})
	buf.Reset()
}

// testSpan is an implementation of [trace.Span] that performs no operations.
type testSpan struct {
	embedded.Span
	isRecording bool
	sc          trace.SpanContext
}

var _ trace.Span = &testSpan{}

func (s *testSpan) SpanContext() trace.SpanContext { return s.sc }

func (s *testSpan) IsRecording() bool { return s.isRecording }

func (s *testSpan) SetStatus(codes.Code, string) {}

func (*testSpan) SetError(bool) {}

func (*testSpan) SetAttributes(...attribute.KeyValue) {}

func (*testSpan) End(...trace.SpanEndOption) {}

func (*testSpan) RecordError(error, ...trace.EventOption) {}

func (*testSpan) AddEvent(string, ...trace.EventOption) {}

func (*testSpan) SetName(string) {}

func (*testSpan) TracerProvider() trace.TracerProvider { return nil }

func (s *testSpan) AddLink(trace.Link) {}
