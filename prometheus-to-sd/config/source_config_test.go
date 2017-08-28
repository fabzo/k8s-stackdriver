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

package config

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fabzo/k8s-stackdriver/prometheus-to-sd/flags"
)

func TestNewSourceConfig(t *testing.T) {
	correct := [...]struct {
		component   string
		host        string
		port        string
		path        string
		match       []string
		whitelisted string
		output      SourceConfig
	}{
		{"testComponent", "localhost", "1234", "", []string{}, "a,b,c,d",
			SourceConfig{
				Component:   "testComponent",
				Host:        "localhost",
				Port:        1234,
				Path:        "/metrics",
				Match:       []string{},
				Whitelisted: []string{"a", "b", "c", "d"},
			},
		},

		{"testComponent", "localhost", "1234", "/federate", []string{"{job=\"prometheus\"}", "{__name__=~\"job:.*\"}"}, "",
			SourceConfig{
				Component:   "testComponent",
				Host:        "localhost",
				Port:        1234,
				Path:        "/federate",
				Match:       []string{"{job=\"prometheus\"}", "{__name__=~\"job:.*\"}"},
				Whitelisted: nil,
			},
		},
	}

	for _, c := range correct {
		res, err := newSourceConfig(c.component, c.host, c.port, c.path, c.match, c.whitelisted)
		if assert.NoError(t, err) {
			assert.Equal(t, c.output, *res)
		}
	}
}

func TestParseSourceConfig(t *testing.T) {
	correct := [...]struct {
		in     flags.Uri
		output SourceConfig
	}{
		{
			flags.Uri{
				Key: "testComponent",
				Val: url.URL{
					Scheme:   "http",
					Host:     "hostname:1234",
					Path:     "/federate",
					RawQuery: "whitelisted=a,b,c,d&match[]={job=\"prometheus\"}&match[]={__name__=~\"job:.*\"}",
				},
			},
			SourceConfig{
				Component:   "testComponent",
				Host:        "hostname",
				Port:        1234,
				Path:        "/federate",
				Match:       []string{"{job=\"prometheus\"}", "{__name__=~\"job:.*\"}"},
				Whitelisted: []string{"a", "b", "c", "d"},
			},
		},
		{
			flags.Uri{
				Key: "testMetricsPathFallback",
				Val: url.URL{
					Scheme:   "http",
					Host:     "hostname:1234",
					Path:     "",
					RawQuery: "whitelisted=a,b,c,d",
				},
			},
			SourceConfig{
				Component:   "testMetricsPathFallback",
				Host:        "hostname",
				Port:        1234,
				Path:        "/metrics",
				Match:       nil,
				Whitelisted: []string{"a", "b", "c", "d"},
			},
		},
	}

	for _, c := range correct {
		res, err := parseSourceConfig(c.in)
		if assert.NoError(t, err) {
			assert.Equal(t, c.output, *res)
		}
	}

	incorrect := [...]flags.Uri{
		{
			Key: "incorrectHost",
			Val: url.URL{
				Scheme:   "http",
				Host:     "hostname[:1234",
				RawQuery: "whitelisted=a,b,c,d",
			},
		},
		{
			Key: "noPort",
			Val: url.URL{
				Scheme:   "http",
				Host:     "hostname",
				RawQuery: "whitelisted=a,b,c,d",
			},
		},
	}

	for _, c := range incorrect {
		_, err := parseSourceConfig(c)
		assert.Error(t, err)
	}
}
