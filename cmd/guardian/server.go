package main

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/gophero/guardian/internal/buildinfo"
	"github.com/gophero/guardian/internal/infra/postgres"
	"github.com/gophero/guardian/internal/log"
	"github.com/gophero/guardian/internal/server"
	"github.com/gophero/guardian/internal/tracing"
	"github.com/grafana/dskit/services"
	"github.com/prometheus/client_golang/prometheus"
)

type ServerCmd struct {
	Postgres postgres.Config `prefix:"postgres." envprefix:"POSTGRES_" embed:""`

	Tracing tracing.Config `prefix:"tracing." envprefix:"TRACING_" embed:""`

	Metrics struct {
		Enabled bool          `help:"Enable prometheus metrics server." name:"enabled" env:"ENABLED" default:"true"`
		Server  server.Config `prefix:"server." envprefix:"SERVER_" embed:""`
	} `prefix:"metrics." envprefix:"METRICS_" embed:""`

	Profiling struct {
		Enabled              bool          `help:"Enable go profiling server." name:"enabled" env:"ENABLED" default:"true"`
		Server               server.Config `prefix:"server." envprefix:"SERVER_" embed:""`
		BlockProfileRate     int           `help:"This controls the fraction of goroutine blocking events that are reported in the blocking profile." name:"block_profile_rate" env:"BLOCK_PROFILE_RATE" default:"20"`
		MutexProfileFraction int           `help:"This controls the fraction of mutex contention events that are reported in the mutex profile. On average 1/rate events are reported." name:"mutex_profile_fraction" env:"MUTEX_PROFILE_FRACTION" default:"20"`
	} `prefix:"profiling." envprefix:"PROFILING_" embed:""`
}

func (cmd *ServerCmd) Run(ctx context.Context, buildInfo buildinfo.BuildInfo) error {
	// Log build information.
	buildInfo.Log(log.Logger)

	// Setup tracing.
	tm, err := tracing.New(cmd.Tracing, buildInfo)
	if err != nil {
		return fmt.Errorf("main: new tracing manager: %w", err)
	}

	if err := tm.Init(ctx); err != nil {
		return fmt.Errorf("main: init tracing manager: %w", err)
	}
	defer tm.Shutdown()

	// Setup postgres.
	pgPool, err := postgres.Connect(ctx, cmd.Postgres)
	if err != nil {
		return fmt.Errorf("main: connect postgres: %w", err)
	}
	defer pgPool.Close()

	prometheus.MustRegister(postgres.NewCollector(pgPool, "primary"))

	// Setup services.
	svc := make([]services.Service, 0)

	if cmd.Metrics.Enabled {
		s, err := server.NewMetricsServer(cmd.Metrics.Server)
		if err != nil {
			return fmt.Errorf("main: new metrics server: %w", err)
		}

		svc = append(svc, s)
	}

	if cmd.Profiling.Enabled {
		runtime.SetBlockProfileRate(cmd.Profiling.BlockProfileRate)
		runtime.SetMutexProfileFraction(cmd.Profiling.MutexProfileFraction)

		s, err := server.NewProfilingServer(cmd.Profiling.Server)
		if err != nil {
			return fmt.Errorf("main: new profiling server: %w", err)
		}

		svc = append(svc, s)
	}

	// Setup manager for services.
	manager, err := services.NewManager(svc...)
	if err != nil {
		return fmt.Errorf("main: new services manager: %w", err)
	}

	if err := manager.StartAsync(ctx); err != nil {
		return fmt.Errorf("main: start manager: %w", err)
	}

	// Wait of all services to be running.
	hCtx, hCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer hCancel()

	if err := manager.AwaitHealthy(hCtx); err != nil {
		return fmt.Errorf("main: await healthy manager: %w", err)
	}

	log.Info().Msg("all services running")

	// Block till all services are stopped.
	if err := manager.AwaitStopped(context.Background()); err != nil {
		return fmt.Errorf("main: await stopped manager: %w", err)
	}

	log.Info().Msg("all services stopped")
	return nil
}
