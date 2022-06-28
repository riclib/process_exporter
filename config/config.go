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

package config

import (
	"io/ioutil"

	pconfig "github.com/prometheus/common/config"
	"gopkg.in/yaml.v2"
)

// Metric contains values that define a metric

//CollectConfig contain selectors for metrics to be collected
type Target struct {
	Endpoint string `yaml:"endpoint"` // TODO: Add any extra field you may need per target
	Token    string `yaml:"token,omitempty"`
}

// Config contains metrics and headers defining a configuration
type Config struct {
	//	ManagedIdentityURL     string                   `yaml:"managed_identity_url"`
	//	AzureURL               string                   `yaml:"azure_url"`
	// TODO: Do you need managed identity? Do you need more config fields?
	Targets          map[string]Target        `yaml:"targets"`
	HTTPClientConfig pconfig.HTTPClientConfig `yaml:"http_client_config,omitempty"`
}

func LoadConfig(configPath string) (Config, error) {
	var config Config
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return config, err
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, err
	}

	// Complete Defaults

	return config, nil
}
