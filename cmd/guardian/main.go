package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/gophero/guardian/pkg/bedrock/buildinfo"
	"github.com/gophero/guardian/pkg/bedrock/log"
)

var (
	program          = "guardian"
	branch           = ""
	buildTimeRFC3339 = ""
)

type Cmd struct {
	Version kong.VersionFlag `help:"Print version and build information." name:"version" short:"v"`
	Config  kong.ConfigFlag  `help:"Configuration file." name:"config" env:"CONFIG"`

	Log log.Config `prefix:"log." envprefix:"LOG_" embed:""`

	Server ServerCmd `cmd:"" help:"Start server."`
}

func main() {
	defer func() {
		if rvr := recover(); rvr != nil {
			log.Fatal().Str(log.Stack(2)).Msg(fmt.Sprintf("panic: %v", rvr))
		}
	}()

	buildInfo, err := buildinfo.New(program, branch, buildTimeRFC3339)
	if err != nil {
		log.Fatal().Err(err).Msg("error buildinfo.New")
	}

	prometheus.MustRegister(buildInfo.Collector())

	var cmd Cmd
	kCtx := kong.Parse(&cmd,
		kong.Name(buildInfo.Program),
		kong.Description("A server and Go library for Authentication, Authorization and User Management."),
		kong.UsageOnError(),
		kong.Configuration(kong.JSON, "guardian.json"),
		kong.Vars{
			"version": buildInfo.String(),
		},
	)

	if err := log.Init(cmd.Log); err != nil {
		log.Fatal().Err(err).Msg("error log.Init")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	kCtx.BindTo(ctx, (*context.Context)(nil))

	if err := kCtx.Run(buildInfo); err != nil {
		log.Fatal().Err(err).Msg("command failed")
	}
}
