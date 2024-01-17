package main

import (
	"net/http"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"

	"github.com/angarium-cloud/keepalived_exporter/collector"
)

func init() {
	prometheus.MustRegister(version.NewCollector("keepalived_exporter"))
}

var Version, commit, date string

func main() {
	var (
		metricsPath = kingpin.Flag(
			"web.telemetry-path",
			"Path under which to expose metrics.",
		).Default("/metrics").String()
		toolkitFlags = webflag.AddFlags(kingpin.CommandLine, ":9650")
		useJSON      = kingpin.Flag("keepalived.use-json", "Send SIGJSON and decode JSON file instead of parsing text files.").Default("false").Bool()
	)

	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("keepalived_exporter"))
	kingpin.Parse()
	logger := promlog.New(promlogConfig)

	level.Info(logger).Log("msg", "Starting keepalived_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())

	coll, err := collector.NewKeepalivedCollector(*useJSON, logger)
	if err != nil {
		panic(err)
	}
	prometheus.MustRegister(coll)

	if *metricsPath != "/" && *metricsPath != "" {
		landingConfig := web.LandingConfig{
			Name:        "Keepalived Exporter",
			Description: "Prometheus Exporter for Keepalived service",
			Version:     version.Info(),
			Links: []web.LandingLinks{
				{
					Address: *metricsPath,
					Text:    "Metrics",
				},
			},
		}
		landingPage, err := web.NewLandingPage(landingConfig)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}
		http.Handle("/", landingPage)
	}
	http.Handle(*metricsPath, promhttp.Handler())
	server := &http.Server{}
	if err := web.ListenAndServe(server, toolkitFlags, logger); err != nil {
		level.Info(logger).Log("err", err)
		os.Exit(1)
	}
}
