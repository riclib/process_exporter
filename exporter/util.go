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
	"context"
	"errors"
	"github.com/prometheus/client_golang/prometheus"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	pconfig "github.com/prometheus/common/config"
	"github.com/riclib/template_exporter/config"
)

type metricInfo struct {
	Desc *prometheus.Desc
	Type prometheus.ValueType
}

type metrics map[string]metricInfo

// Joins the parts separated by underscore to make a sesible metric name
func MakeMetricName(parts ...string) string {
	return strings.Join(parts, "_")
}

func btof(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}
func FetchJson(ctx context.Context, logger log.Logger, endpoint string, config config.Config) ([]byte, error) {
	httpClientConfig := config.HTTPClientConfig
	client, err := pconfig.NewClientFromConfig(httpClientConfig, "fetch_json", true, false)
	if err != nil {
		level.Error(logger).Log("msg", "Error generating HTTP client", "err", err) //nolint:errcheck
		return nil, err
	}
	req, err := http.NewRequest("GET", endpoint, nil)
	req = req.WithContext(ctx)
	if err != nil {
		level.Error(logger).Log("msg", "Failed to create request", "err", err) //nolint:errcheck
		return nil, err
	}

	if req.Header.Get("Accept") == "" {
		req.Header.Add("Accept", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
			level.Error(logger).Log("msg", "Failed to discard body", "err", err) //nolint:errcheck
		}
		resp.Body.Close()
	}()

	if resp.StatusCode/100 != 2 {
		return nil, errors.New(resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
