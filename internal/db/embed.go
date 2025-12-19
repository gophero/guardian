package db

import (
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed sql/schema
var sqlSchemaFS embed.FS

func sourceDriver() (source.Driver, error) {
	s, err := iofs.New(sqlSchemaFS, "sql/schema") // no need to s.Close() since fs is embed.FS.
	if err != nil {
		return nil, fmt.Errorf("db: new iofs: %w", err)
	}
	return s, nil
}
