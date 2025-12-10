package tracing

import (
	"context"
	"fmt"

	"github.com/gophero/guardian/internal/buildinfo"
	"github.com/gophero/guardian/internal/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace/noop"
)

// Manager handles configuration, registration and shutdown of global tracer provider.
type Manager struct {
	config Config
	bi     buildinfo.BuildInfo
	client otlptrace.Client
	tp     *trace.TracerProvider
}

// New constructs new [Manager].
func New(c Config, bi buildinfo.BuildInfo) (*Manager, error) {
	client, err := c.client()
	if err != nil {
		return nil, err
	}

	return &Manager{
		config: c,
		bi:     bi,
		client: client,
	}, nil
}

// Init initialize the tracer provider.
func (s *Manager) Init(ctx context.Context) error {
	if !s.config.Enabled {
		otel.SetTracerProvider(noop.TracerProvider{})
		log.Info().Msg("tracing provider disbaled")
		return nil
	}

	exporter, err := otlptrace.New(ctx, s.client)
	if err != nil {
		return fmt.Errorf("tracing: new otlptrace: %w", err)
	}

	res, err := resource.New(
		ctx,
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(s.bi.Program),
			semconv.ServiceVersionKey.String(s.bi.Version),
		),
		resource.WithProcessRuntimeDescription(),
		resource.WithTelemetrySDK(),
	)
	if err != nil {
		return fmt.Errorf("tracing: new resource: %w", err)
	}

	s.tp = trace.NewTracerProvider(trace.WithBatcher(exporter), trace.WithResource(res))

	otel.SetTracerProvider(s.tp)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(cause error) {
		log.Err(err).Msg("otel handler error")
	}))

	log.Info().Msg("tracing provider configured")

	return nil
}

// Shutdown shuts down tracer provider.
func (s *Manager) Shutdown() {
	if !s.config.Enabled {
		return
	}

	if err := s.tp.Shutdown(context.Background()); err != nil {
		log.Err(err).Msg("tracing provider shutdown failed")
	}
	log.Info().Msg("tracing provider shutdown successful")
}
