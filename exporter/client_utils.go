package exporter

import (
	"bytes"
	"errors"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	pconfig "github.com/prometheus/common/config"
	"github.com/riclib/template_exporter/config"
	"io"
	"io/ioutil"
	"net/http"
)

/* Support for managed Identity */

//TODO: Move to package

type ManagedIdentityResponse struct {
	AccessToken  string `json:"access_token"`
	ClientId     string `json:"client_id"`
	ExpiresIn    string `json:"expires_in"`
	ExpiresOn    string `json:"expires_on"`
	ExtExpiresIn string `json:"ext_expires_in"`
	NotBefore    string `json:"not_before"`
	Resource     string `json:"resource"`
	TokenType    string `json:"token_type"`
}

func FetchJsonWithTokenGET(logger log.Logger, config config.Config, endpoint string, token string) ([]byte, error) {
	httpClientConfig := config.HTTPClientConfig
	client, err := pconfig.NewClientFromConfig(httpClientConfig, "fetch_json", true, false)
	if err != nil {
		level.Error(logger).Log("msg", "Error generating HTTP client", "err", err) //nolint:errcheck
		return nil, err
	}
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		level.Error(logger).Log("msg", "Failed to create request", "err", err) //nolint:errcheck
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

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

func FetchJsonWithTokenPOST(logger log.Logger, config config.Config, endpoint string, token string, body []byte) ([]byte, error) {
	httpClientConfig := config.HTTPClientConfig
	client, err := pconfig.NewClientFromConfig(httpClientConfig, "fetch_json", true, false)
	if err != nil {
		level.Error(logger).Log("msg", "Error generating HTTP client", "err", err) //nolint:errcheck
		return nil, err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		level.Error(logger).Log("msg", "Failed to create request", "err", err) //nolint:errcheck
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

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
