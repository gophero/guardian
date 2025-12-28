package server

import (
	"net/http"
	"net/http/pprof"
)

// NewProfilingServer creates a new [Server] with net/http/pprof handlers.
func NewProfilingServer(config Config) (*Server, error) {
	if config.Network == "tcp" && config.Addr == "" {
		config.Addr = "localhost:9003"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /debug/pprof/", pprof.Index)
	mux.HandleFunc("GET /debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("GET /debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("GET /debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("GET /debug/pprof/trace", pprof.Trace)

	s, err := newServer("profiling", config, mux)
	if err != nil {
		return nil, err
	}

	return s, nil
}
