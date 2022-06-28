// Copyright 2020 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	cache "github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/riclib/template_exporter/config"
	exporter "github.com/riclib/template_exporter/exporter"
)

var (
	configFile    = kingpin.Flag("config.file", "TEMPLATE exporter configuration file.").ExistingFile()
	logRunsFile   = kingpin.Flag("lokioutput.file", "File to write loki output to").String()
	listenAddress = kingpin.Flag("web.listen-address", "The address to listen on for HTTP requests.").Default(":19098").String()
	configCheck   = kingpin.Flag("config.check", "If true validate the config file and then exit.").Default("false").Bool()
	target        = kingpin.Flag("target", "The address to listen on for HTTP requests.").String()
)

func Run() {

	promlogConfig := &promlog.Config{}

	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	//	kingpin.Version(version.Print("template_exporter")) TODO: figure out how to use, hint: https://github.com/prometheus/common/blob/master/version/info.go
	kingpin.Version("template_exporter 1.0.0")
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(promlogConfig)
	var logRuns log.Logger
	if *logRunsFile == "" {
		level.Info(logger).Log("msg", "running without logging metrics to lokioutput") //nolint:errcheck
		logRuns = nil
	} else {
		f, err := os.OpenFile(*logRunsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			level.Error(logger).Log("msg", "could not create loki output file") //nolint:errcheck
			os.Exit(1)
		}
		logRuns = log.NewJSONLogger(log.NewSyncWriter(f))
		logRuns = log.With(logRuns, "ts", log.DefaultTimestampUTC)
	}

	level.Info(logger).Log("msg", "Starting template_exporter", "version", version.Info()) //nolint:errcheck
	level.Info(logger).Log("msg", "Build context", "build", version.BuildContext())        //nolint:errcheck

	level.Info(logger).Log("msg", "Loading config file", "file", *configFile) //nolint:errcheck
	config, err := config.LoadConfig(*configFile)
	if err != nil {
		level.Error(logger).Log("msg", "Error loading config", "err", err) //nolint:errcheck
		os.Exit(1)
	}
	level.Info(logger).Log("msg", "Loaded config file", "config") //nolint:errcheck
	if *configCheck {
		os.Exit(0)
	}

	runsCache := make(map[string]*cache.Cache) //one per environment
	for name, _ := range config.Targets {
		runsCache[name] = cache.New(time.Hour+48, time.Hour*2)
	}

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/probe", func(w http.ResponseWriter, req *http.Request) {
		probeHandler(w, req, logger, config, &runsCache, logRuns)
	})
	level.Info(logger).Log("msg", "Starting server", "listenAddress", *listenAddress) //nolint:errcheck

	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		level.Error(logger).Log("msg", "Failed to start the server", "err", err) //nolint:errcheck
	}

}

func probeHandler(w http.ResponseWriter, r *http.Request, logger log.Logger, config config.Config, templateCaches *map[string]*cache.Cache, exportLogger log.Logger) {

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	r = r.WithContext(ctx)

	registry := prometheus.NewPedanticRegistry()
	env := r.URL.Query().Get("target")
	if env == "" {
		const e = "Target parameter is missing"
		http.Error(w, e, http.StatusBadRequest)
		level.Error(logger).Log("msg", e) //nolint:errcheck
		return
	}
	target, found := config.Targets[env]
	if !found {
		e := "Target '" + env + "' not in config"
		http.Error(w, e, http.StatusBadRequest)
		level.Error(logger).Log("msg", "env not in config", "env", env) //nolint:errcheck
		return
	}
	templateMetricCollector := exporter.TemplateMetricCollector{}
	templateMetricCollector.Env = env
	templateMetricCollector.Logger = log.With(logger, "env", env)
	templateMetricCollector.Config = config
	templateMetricCollector.Target = target

	templateMetricCollector.Cache = (*templateCaches)[env]
	templateMetricCollector.LokiExportLogger = exportLogger

	// Collect primary node topology

	registry.MustRegister(templateMetricCollector)
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)

}
