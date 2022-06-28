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

package exporter

import (
	"github.com/go-kit/kit/log"
	cache "github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/riclib/template_exporter/config"
)

const namespace = "template" //TODO: Set namespace

type TemplateMetricCollector struct {

	// Add here any extra context fields

	Logger           log.Logger
	LokiExportLogger log.Logger
	Config           config.Config
	Cache            *cache.Cache
	Target           config.Target
	Env              string
}

// Describe is a
func (mc TemplateMetricCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range pipelineRunMetrics {
		ch <- m.Desc
	}
}

// Collect all metrics
func (mc TemplateMetricCollector) Collect(ch chan<- prometheus.Metric) {

	mc.CollectTemplateMetrics(ch) //TODO: Call the different exporters you created from export_template.go

}
