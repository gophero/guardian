package log

import (
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

// tracingHook extract [trace.TraceID] and [trace.SpanID] from ctx and inject it in logs.
type tracingHook struct{}

var _ zerolog.Hook = tracingHook{}

// Run implements [zerolog.Hook].
func (h tracingHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	ctx := e.GetCtx()
	span := trace.SpanFromContext(ctx)

	if span.IsRecording() {
		spanCtx := span.SpanContext()

		if spanCtx.HasTraceID() {
			e.Str("trace_id", spanCtx.TraceID().String())
		}

		if spanCtx.HasSpanID() {
			e.Str("span_id", spanCtx.SpanID().String())
		}
	}
}
