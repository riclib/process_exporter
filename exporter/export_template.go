package exporter

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
)

//TODO: Use this file asa template for each kind of export

const templateSubsystem = "pipelineruns" //TODO: Set to the subsystem you're monitoring. metrics will be <namespace>_<subsystem>_metric
// Each file you copy should have a different subsystem

type QueryInput struct { // TODO: Set to expected input
	LastUpdatedAfter  time.Time `json:"lastUpdatedAfter"`
	LastUpdatedBefore time.Time `json:"lastUpdatedBefore"`
	ContinuationToken string    `json:"continuationToken"`
}
type ResultRow struct {
	Id          string `json:"id"`
	MetricValue int    `json:"metric_value"`
}
type QueryResult struct {
	Value             []ResultRow `json:"value"`
	ContinuationToken string      `json:"continuationToken"` //template
}

var (
	templateLabelNames = []string{"id"} // TODO: Add more labels
)

var pipelineRunMetrics = metrics{ // TODO: Add more metrics
	"template": newTemplateMetric("template", "Example: Shows length of collected json with id 1", prometheus.GaugeValue, templateLabelNames),
}

// Maps strings returned about status to numbers, so we can log in prometheus

var templateMapping = map[string]float64{ //TODO: Are the any textual metrics you meed to map to a number?
	"Queued":     1,
	"InProgress": 2,
	"Canceling":  5,
}

func newTemplateMetric(metricName string, docString string, t prometheus.ValueType, labelNames []string) metricInfo {
	return metricInfo{
		Desc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, templateSubsystem, metricName),
			docString,
			labelNames,
			nil,
		),
		Type: t,
	}
}

func (mc TemplateMetricCollector) CollectTemplateMetrics(ch chan<- prometheus.Metric) {

	logger := log.With(mc.Logger, "subsystem", templateSubsystem)
	endpoint := mc.Target.Endpoint
	token := mc.Target.Token
	continuationToken := ""

	for {

		//Snippet for input parameters
		/*params := QueryInput{
			ContinuationToken: continuationToken, // TODO: replace by whatever paging logic you may need
		}

		bodyBytes, err := json.Marshal(params)
		if err != nil {
			level.Error(logger).Log("msg", "couldn't create params for collection", "err", err) //nolint:errcheck
			return
		}
		level.Debug(logger).Log("msg", "query parameters", "params", bodyBytes) //nolint:errcheck

			response, err := FetchJsonWithTokenPOST(logger, mc.Config, endpoint, mc.Target.Token, bodyBytes)
			if err != nil {
				level.Error(logger).Log("msg", "couldn't fetch json", "err", err) //nolint:errcheck
				return
			} */

		response, err := FetchJsonWithTokenGET(logger, mc.Config, endpoint, token)
		if err != nil {
			level.Error(logger).Log("msg", "couldn't fetch json", "err", err) //nolint:errcheck
			return
		}

		simplestLabels := []string{"1"}
		ch <- prometheus.MustNewConstMetric(pipelineRunMetrics["template"].Desc, prometheus.GaugeValue, float64(len(response)), simplestLabels...)

		var result QueryResult

		if err := json.Unmarshal(response, &result); err != nil {
			level.Error(logger).Log("msg", "couldn't parse json", "err", err) //nolint:errcheck
			return
		}

		level.Debug(logger).Log("msg", "Parsed results", "len", len(result.Value)) //nolint:errcheck

		for _, v := range result.Value {

			reportMetric := false // This allows to filter metric based on any criteria

			// Decide if we want to report metric
			if v.Id != "" {
				reportMetric = true
				level.Debug(logger).Log("msg", "Decided to log metric", "id", v.Id) //nolint:errcheck
			}

			if reportMetric {

				// Write it to the log if it is a new id TODO: Implement any decision needed for logging
				if mc.LokiExportLogger != nil {
					var cachedItem interface{}
					cachedItem, found := mc.Cache.Get(v.Id)
					if !found {
						mc.Cache.Add(v.Id, v, cache.DefaultExpiration)
						logTemplate(logger, mc.LokiExportLogger, v, mc.Env)
					} else {
						cachedRun := cachedItem.(ResultRow)
						if cachedRun.MetricValue != v.MetricValue {
							logTemplate(logger, mc.LokiExportLogger, v, mc.Env)
							mc.Cache.Replace(v.Id, v, cache.DefaultExpiration)
						}
					}
				}

				labels := []string{
					v.Id,
				}
				ch <- prometheus.MustNewConstMetric(pipelineRunMetrics["template"].Desc, prometheus.GaugeValue, float64(v.MetricValue), labels...)

				// TODO: Use commented snippet below to translate any string statuses to metrics
				/*				status, found := templateMapping[v.Id] // this snippet implements translating strings to metrics such as statuses
								if !found {
									level.Error(logger).Log("msg", "Found unexpected status", "status", status) //nolint:errcheck
									status = -1.0
								ch <- prometheus.MustNewConstMetric(pipelineRunMetrics["status"].Desc, prometheus.GaugeValue, status, labels...)
								}
				*/
			}
		}
		continuationToken = result.ContinuationToken // TODO: Implement any paging logic to decide if finish paging
		if continuationToken == "" {
			break
		}

	}

}

func logTemplate(debugLogger log.Logger, outputLogger log.Logger, run ResultRow, env string) error {

	if outputLogger != nil {
		return outputLogger.Log( // TODO: Implement any logic needed for logging new metrics
			"msg", "Run",
			"id", run.Id,
			"value", run.MetricValue,
		)
	} else {
		return errors.New("Call to log with null logger")
	}
}
