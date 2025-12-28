package migration

import (
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source"
)

type Factory interface {
	NewMigrate(sourceDriver string, src source.Driver, table string) (*migrate.Migrate, error)
}
