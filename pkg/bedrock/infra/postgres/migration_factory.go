package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/rs/zerolog"

	"github.com/gophero/guardian/pkg/bedrock/log"
	"github.com/gophero/guardian/pkg/migration"
)

type MigrationConfig struct {
	StatementTimeout time.Duration `help:"Timeout per statement. Zero means no timeout." name:"statement_timeout" env:"STATEMENT_TIMEOUT" default:"0"`
	LockTimeout      time.Duration `help:"Maximum wait time for acquiring database lock." name:"lock_timeout" env:"LOCK_TIMEOUT" default:"15s"`
}

type migrationFactory struct {
	logger zerolog.Logger
	config MigrationConfig
	db     *sql.DB
}

var _ migration.Factory = (*migrationFactory)(nil)

func NewMigrationFactory(db *sql.DB, config MigrationConfig) (*migrationFactory, error) {
	if config.LockTimeout < 0 {
		return nil, errors.New("migrate: LockTimeout cannot be zero or negative")
	}

	return &migrationFactory{
		logger: log.With().Str("logger", "migrator").Logger(),
		config: config,
		db:     db,
	}, nil
}

func (f *migrationFactory) NewMigrate(sourceDriver string, src source.Driver, table string) (*migrate.Migrate, error) {
	pgxDriver, err := pgx.WithInstance(f.db, &pgx.Config{
		MigrationsTable:  table,
		StatementTimeout: f.config.StatementTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("migrate: new pgx migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance(sourceDriver, src, "pgx5", pgxDriver)
	if err != nil {
		return nil, fmt.Errorf("migrate: new migrate: %w", err)
	}

	m.LockTimeout = f.config.LockTimeout
	m.Log = newMigrateLog(f.logger.With().Str("migration_table", table).Logger())

	return m, nil
}

type migrateLog struct {
	logger zerolog.Logger
}

var _ migrate.Logger = migrateLog{}

func newMigrateLog(l zerolog.Logger) migrateLog {
	return migrateLog{logger: l}
}

// Printf implements [migrate.Logger].
func (l migrateLog) Printf(format string, v ...any) {
	msg := strings.TrimSpace(fmt.Sprintf(format, v...))

	if strings.HasPrefix(format, "error:") { // https://github.com/golang-migrate/migrate/blob/257fa847d614efe3948c25e9033e92b930527dec/migrate.go#L975
		l.logger.Error().Msg(msg)
		return
	}

	l.logger.Info().Msg(msg)
}

// Verbose implements [migrate.Logger].
func (l migrateLog) Verbose() bool {
	return false
}
