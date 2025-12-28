package db

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"

	"github.com/gophero/guardian/pkg/migration"
)

// Migration table name.
var MigrationsTable = "gophero_guardian_migrations"

func RunMigrations(f migration.Factory) error {
	src, err := sourceDriver()
	if err != nil {
		return err
	}

	m, err := f.NewMigrate("iofs", src, MigrationsTable)
	if err != nil {
		return fmt.Errorf("db: new migrate: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
