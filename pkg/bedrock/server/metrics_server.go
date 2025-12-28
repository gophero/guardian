package server

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewMetricsServer creates a new [Server] with prometheus handler.
func NewMetricsServer(config Config) (*Server, error) {
	if config.Network == "tcp" && config.Addr == "" {
		config.Addr = "localhost:9002"
	}

	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.Handler())

	s, err := newServer("metrics", config, mux)
	if err != nil {
		return nil, err
	}

	return s, nil
}
