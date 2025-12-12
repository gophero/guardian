package main

import (
	"context"
	"errors"

	"github.com/gophero/guardian/internal/buildinfo"
)

type MigrationCmd struct{}

func (cmd *MigrationCmd) Run(ctx context.Context, bi buildinfo.BuildInfo) error {
	return errors.New("command not implemented")
}
