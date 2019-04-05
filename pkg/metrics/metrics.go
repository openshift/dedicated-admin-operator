// Copyright 2018 RedHat
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"net/http"

	"github.com/openshift/dedicated-admin-operator/config"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	MetricsEndpoint = ":8080"
)

var (
	metricBlacklistedGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "dedicated_admin_blacklisted_projects",
		Help: "Report how many namespaces have been blacklisted for dedicated-admin-operator",
	}, []string{"name"})
)

// StartMetrics register metrics and exposes them
func StartMetrics() {
	// Register metrics and start serving them on /metrics endpoint
	RegisterMetrics()
	http.Handle("/metrics", prometheus.Handler())
	go http.ListenAndServe(MetricsEndpoint, nil)
}

// RegisterMetrics for the operator
func RegisterMetrics() error {
	err := prometheus.Register(metricBlacklistedGauge)

	return err
}

// IncEventGauge will increment a gauge and set appropriate labels.
func incEventGauge(gauge *prometheus.GaugeVec) {
	gauge.With(prometheus.Labels{"name": config.OperatorName}).Inc()
}

// UpdateBlacklistedGauge sets the gauge metric with the number of blacklisted projects
func UpdateBlacklistedGauge(blacklistedProjects map[string]bool) {
	metricBlacklistedGauge.Reset()
	for _, isBlacklisted := range blacklistedProjects {
		if isBlacklisted {
			incEventGauge(metricBlacklistedGauge)
		}
	}
}
