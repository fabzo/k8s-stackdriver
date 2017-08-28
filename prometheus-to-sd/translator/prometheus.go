/*
Copyright 2017 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package translator

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

const matchParameter = "match[]="

// GetPrometheusMetrics scrapes metrics from the given host and port using /metrics handler.
func GetPrometheusMetrics(host string, port uint, path string, match []string) (map[string]*dto.MetricFamily, error) {
	url := fmt.Sprintf("http://%s:%d/%s", host, port, path) + createMatchQueryString(match)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request %s failed: %v", url, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body - %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed - %q, response: %q", resp.Status, string(body))
	}

	data := string(body)

	parser := &expfmt.TextParser{}
	return parser.TextToMetricFamilies(strings.NewReader(data))
}

func createMatchQueryString(match []string) string {
	if match == nil || len(match) <= 0 {
		return ""
	}

	return "?" + matchParameter + strings.Join(match, "&" + matchParameter)
}