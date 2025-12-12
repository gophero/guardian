package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gophero/guardian/internal/log"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// Connect creates new [pgxpool.Pool].
func Connect(ctx context.Context, config Config) (*pgxpool.Pool, error) {
	poolConf, err := config.parse()
	if err != nil {
		return nil, err
	}

	dbPool, err := pgxpool.NewWithConfig(ctx, poolConf)
	if err != nil {
		return nil, fmt.Errorf("postgres: new pgx pool: %w", err)
	}

	if err := dbPool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("postgres: ping db: %w", err)
	}

	log.Info().Msg("connected to postgres")

	return dbPool, nil
}

// StdConnect creates new [sql.DB]. The only use case for this is to run migration using goose. This ingores all pool related configurations.
func StdConnect(ctx context.Context, config Config) (*sql.DB, error) {
	poolConf, err := config.parse()
	if err != nil {
		return nil, err
	}

	connStr := stdlib.RegisterConnConfig(poolConf.ConnConfig)

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("postgres: open sql db: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("postgres: ping db: %w", err)
	}

	log.Info().Msg("connected to postgres")

	return db, nil
}
