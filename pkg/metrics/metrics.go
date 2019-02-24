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

	"github.com/prometheus/client_golang/prometheus"
)

const (
	MetricsEndpoint = ":8080"
)

var (
	metricBlacklistedCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "dedicated_admin_blacklisted",
		Help: "Report how many namespaces have been blacklisted and did not receive the rolebinding",
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
	err := prometheus.Register(metricBlacklistedCounter)

	return err
}

// IncBlacklistedCount increments the counter for black listed namespaces
func IncBlacklistedCount() {
	metricBlacklistedCounter.With(prometheus.Labels{"name": "dedicated-admin-operator"}).Inc()
}
