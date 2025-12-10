package tracing

import (
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
)

type Config struct {
	Enabled     bool              `help:"Enable tracing." name:"enabled" env:"ENABLED" default:"true"`
	Client      string            `help:"HTTP or gRPC client." name:"client" env:"CLIENT" enum:"http,grpc" default:"grpc"`
	EndpointURL string            `help:"Target URL to which the exporter is going to send traces." name:"endpoint_url" env:"ENDPOINT_URL" default:""`
	Headers     map[string]string `help:"Key-value pairs to be used as headers associated with gRPC or HTTP requests." name:"headers" env:"HEADERS" mapsep:","`
	Compression string            `help:"Compression algorithm." name:"compression" env:"COMPRESSION" enum:"gzip," default:""`
	Timeout     time.Duration     `help:"Maximum time the OTLP exporter will wait for each batch export." name:"timeout" env:"TIMEOUT" default:"10s"`
}

func (c Config) client() (otlptrace.Client, error) {
	if !c.Enabled {
		return nil, nil
	}

	switch c.Client {
	case "grpc":
		opts := make([]otlptracegrpc.Option, 0)

		if c.EndpointURL != "" {
			opts = append(opts, otlptracegrpc.WithEndpointURL(c.EndpointURL))
		}

		if c.Compression != "" {
			opts = append(opts, otlptracegrpc.WithCompressor(c.Compression))
		}

		if len(c.Headers) > 0 {
			opts = append(opts, otlptracegrpc.WithHeaders(c.Headers))
		}

		if c.Timeout != 0 {
			opts = append(opts, otlptracegrpc.WithTimeout(c.Timeout))
		}

		return otlptracegrpc.NewClient(opts...), nil
	case "http":
		opts := make([]otlptracehttp.Option, 0)

		if c.EndpointURL != "" {
			opts = append(opts, otlptracehttp.WithEndpointURL(c.EndpointURL))
		}

		if c.Compression == "gzip" {
			opts = append(opts, otlptracehttp.WithCompression(otlptracehttp.GzipCompression))
		}

		if len(c.Headers) > 0 {
			opts = append(opts, otlptracehttp.WithHeaders(c.Headers))
		}

		if c.Timeout != 0 {
			opts = append(opts, otlptracehttp.WithTimeout(c.Timeout))
		}

		return otlptracehttp.NewClient(opts...), nil
	default:
		return nil, fmt.Errorf("tracing: `%s` is not a valid client option", c.Client)
	}
}
