package buildinfo

import (
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type BuildInfo struct {
	Program   string
	Version   string
	Time      time.Time // Time at which binary was build.
	Branch    string
	Revision  string // The revision identifier for the current commit or checkout.
	Dirty     bool   // Whether the source tree had local modifications.
	Tags      string // Build tags.
	GoVersion string
	GOOS      string
	GOARCH    string
	Platform  string // GOOS + GOARCH.
}

// Create new BuildInfo from given arguments, buildTime should be formatted according to RFC3339 (time.RFC3339).
func New(program string, branch string, buildTime string) (BuildInfo, error) {
	var bTime time.Time

	if buildTime != "" {
		t, err := time.Parse(time.RFC3339, buildTime)
		if err != nil {
			return BuildInfo{}, fmt.Errorf("buildinfo: parse time format RFC3339: %w", err)
		}

		bTime = t
	}

	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return BuildInfo{}, errors.New("buildinfo: error reading build information embedded in the binary")
	}

	var revision string
	var tags string
	var dirty bool

	for _, v := range bi.Settings {
		if v.Key == "vcs.revision" {
			revision = v.Value
		}

		if v.Key == "vcs.modified" {
			if v.Value == "true" {
				dirty = true
			}
		}

		if v.Key == "-tags" {
			tags = v.Value
		}
	}

	return BuildInfo{
		Program:   program,
		Version:   bi.Main.Version,
		Time:      bTime,
		Branch:    branch,
		Revision:  revision,
		Dirty:     dirty,
		Tags:      tags,
		GoVersion: bi.GoVersion,
		GOOS:      runtime.GOOS,
		GOARCH:    runtime.GOARCH,
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}, nil
}

func (bi BuildInfo) String() string {
	return fmt.Sprintf(`program: %s
version: %s
build_time: %s
branch: %s
revision: %s
dirty: %s
tags: %s
go_version: %s
platform: %s`,
		bi.Program,
		bi.Version,
		bi.Time.Format(time.RFC3339),
		bi.Branch,
		bi.Revision,
		strconv.FormatBool(bi.Dirty),
		bi.Tags,
		bi.GoVersion,
		bi.Platform,
	)
}

// LogAttrs returns key, value pair of BuildInfo to be passed slog.Logger for logging.
func (bi BuildInfo) LogAttrs() []any {
	return []any{
		"program", bi.Program,
		"version", bi.Version,
		"build_time", bi.Time,
		"branch", bi.Branch,
		"revision", bi.Revision,
		"dirty", bi.Dirty,
		"tags", bi.Tags,
		"go_version", bi.GoVersion,
		"platform", bi.Platform,
	}
}

// Collector create a prometheus metric with a constant '1' value labeled by build information.
func (bi BuildInfo) Collector() prometheus.Collector {
	return prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: bi.Program,
			Name:      "build_info",
			Help:      "A metric with a constant '1' value labeled by build information.",
			ConstLabels: prometheus.Labels{
				"version":    bi.Version,
				"build_time": bi.Time.Format(time.RFC3339),
				"branch":     bi.Branch,
				"revision":   bi.Revision,
				"dirty":      strconv.FormatBool(bi.Dirty),
				"tags":       bi.Tags,
				"goversion":  bi.GoVersion,
				"goos":       bi.GOOS,
				"goarch":     bi.GOARCH,
			},
		},
		func() float64 { return 1 },
	)
}
