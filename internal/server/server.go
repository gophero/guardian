package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/gophero/guardian/internal/log"
	"github.com/grafana/dskit/services"
	"github.com/rs/zerolog"
)

type Config struct {
	Network string `help:"TCP or Unix domain socket." name:"network" env:"NETWORK" enum:"tcp,unix" default:"tcp"`
	Addr    string `help:"The address on which http server will listen. Path to a file in case of unix domain socket." name:"addr" env:"ADDR"`
	H2C     bool   `help:"Use unencrypted h2c form of http/2" name:"h2c" env:"H2C" default:"false"`
}

// Server configures and wraps [http.Server] as [services.Service].
type Server struct {
	*services.BasicService

	config  Config
	logger  zerolog.Logger
	httpSrv *http.Server
	errChan chan error
}

func newServer(name string, config Config, h http.Handler) (*Server, error) {
	if config.Addr == "" {
		return nil, errors.New("server: Addr cannot be empty")
	}

	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)

	if config.H2C {
		protocols.SetUnencryptedHTTP2(true)
	} else {
		protocols.SetHTTP2(true)
	}

	logger := log.Logger.With().Str("server", name).Logger()

	httpSrv := &http.Server{
		Handler:   h,
		ErrorLog:  log.NewStdLog(logger, zerolog.ErrorLevel),
		Protocols: protocols,
	}

	s := &Server{config: config, logger: logger, httpSrv: httpSrv, errChan: make(chan error)}
	s.BasicService = services.NewBasicService(s.start, s.running, s.stop)

	return s, nil
}

func (s *Server) start(ctx context.Context) error {
	network, addr := s.config.Network, s.config.Addr
	ln, err := net.Listen(network, addr)
	if err != nil {
		return fmt.Errorf("server: listening for network `%s` at address `%s`: %w", network, addr, err)
	}

	s.logger.Info().
		Ctx(ctx).
		Str("network", network).
		Str("addr", addr).
		Str("protocols", s.httpSrv.Protocols.String()).
		Msg("server started")

	go func() {
		s.errChan <- s.httpSrv.Serve(ln)
	}()

	return nil
}

func (s *Server) running(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return nil
	case err := <-s.errChan:
		return err
	}
}

func (s *Server) stop(reason error) error {
	s.logger.Err(reason).Msg("server shutdown started")

	if err := s.httpSrv.Shutdown(context.Background()); err != nil {
		s.logger.Err(err).Msg("server shutdown failed")
		return err
	}

	s.logger.Info().Msg("server shutdown successful")
	return nil
}
