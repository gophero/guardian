package buildinfo

import (
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

// BuildInfo represent build information of the binary.
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

// New constructs new [BuildInfo] from given arguments. The buildTime string should be in format [time.RFC3339].
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

// String formats build information as a string.
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

// Log logs build information using provided logger.
func (bi BuildInfo) Log(log zerolog.Logger) {
	log.Info().
		CallerSkipFrame(1).
		Str("program", bi.Program).
		Str("version", bi.Version).
		Time("build_time", bi.Time).
		Str("branch", bi.Branch).
		Str("revision", bi.Revision).
		Bool("dirty", bi.Dirty).
		Str("tags", bi.Tags).
		Str("go_version", bi.GoVersion).
		Str("platform", bi.Platform).
		Msg("build information")
}

// Collector creates a prometheus metric with a constant '1' value labeled by build information.
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
