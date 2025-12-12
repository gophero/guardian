package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gophero/guardian/internal/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	URI                   string        `help:"Postgres Connection URI." name:"uri" env:"URI" default:""`
	MaxConns              int32         `help:"Maximum size of the pool." name:"max_conns" env:"MAX_CONNS" default:"20"`
	MinConns              int32         `help:"Minimum size of the pool." name:"min_conns" env:"MIN_CONNS" default:"4"`
	MinIdleConns          int32         `help:"Minimum number of idle connections in the pool." name:"min_idle_conns" env:"MIN_IDLE_CONNS" default:"0"`
	MaxConnLifetime       time.Duration `help:"Duration since creation after which a connection will be automatically closed." name:"max_conn_lifetime" env:"MAX_CONN_LIFETIME" default:"1h"`
	MaxConnIdleTime       time.Duration `help:"Duration after which an idle connection will be automatically closed by the health check." name:"max_conn_idle_time" env:"MAX_CONN_IDLE_TIME" default:"30m"`
	HealthCheckPeriod     time.Duration `help:"Duration between checks of the health of idle connections." name:"health_check_period" env:"HEALTH_CHECK_PERIOD" default:"1m"`
	MaxConnLifetimeJitter time.Duration `help:"Duration after MaxConnLifetime to randomly decide to close a connection. This helps prevent all connections from being closed at the exact same time, starving the pool." name:"max_conn_lifetime_jitter" env:"MAX_CONN_LIFETIME_JITTER" default:"10m"`
}

func (c Config) parse() (*pgxpool.Config, error) {
	if c.URI == "" {
		return nil, errors.New("postgres: URI cannot be empty")
	}

	if c.MinConns < 0 {
		return nil, errors.New("postgres: MinConns cannot be negative")
	}

	if c.MinIdleConns < 0 {
		return nil, errors.New("postgres: MinIdleConns cannot be negative")
	}

	if c.MaxConns == 0 || c.MaxConns < c.MinConns || c.MaxConns < c.MinIdleConns {
		return nil, errors.New("postgres: MaxConns cannot be zero or, less than MinConns or MinIdleConns")
	}

	if c.MaxConnLifetime <= 0 {
		return nil, errors.New("postgres: MaxConnLifetime cannot be zero or negative")
	}

	if c.MaxConnIdleTime <= 0 {
		return nil, errors.New("postgres: MaxConnIdleTime cannot be zero or negative")
	}

	if c.HealthCheckPeriod <= 0 {
		return nil, errors.New("postgres: HealthCheckPeriod cannot be zero or negative")
	}

	if c.MaxConnLifetimeJitter < 0 {
		return nil, errors.New("postgres: MaxConnLifetimeJitter cannot be negative")
	}

	poolConf, err := pgxpool.ParseConfig(c.URI)
	if err != nil {
		return nil, fmt.Errorf("postgres: parse pgx pool config: %w", err)
	}

	poolConf.MaxConns = c.MaxConns
	poolConf.MinConns = c.MinConns
	poolConf.MinIdleConns = c.MinIdleConns
	poolConf.MaxConnLifetime = c.MaxConnLifetime
	poolConf.MaxConnIdleTime = c.MaxConnIdleTime
	poolConf.HealthCheckPeriod = c.HealthCheckPeriod
	poolConf.MaxConnLifetimeJitter = c.MaxConnLifetimeJitter

	// Load custom defined types.
	poolConf.AfterConnect = func(ctx context.Context, c *pgx.Conn) error {
		types, err := c.LoadTypes(ctx, []string{
			"citext",
		})
		if err != nil {
			log.Error().Err(err).Msg("failed to load types for postgres connection")
			return err
		}
		c.TypeMap().RegisterTypes(types)
		return nil
	}

	return poolConf, nil
}
