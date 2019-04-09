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

package operator

import (
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	operatorconfig "github.com/openshift/dedicated-admin-operator/config"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var Services = map[string]corev1.Service{
	"dedicated-admin-operator": {
		ObjectMeta: metav1.ObjectMeta{
			Name:      operatorconfig.OperatorName,
			Namespace: operatorconfig.OperatorNamespace,
			Labels: map[string]string{
				"k8s-app": operatorconfig.OperatorName,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "metrics",
					Port: 8080,
				},
			},
			Selector: map[string]string{
				"k8s-app": operatorconfig.OperatorName,
			},
			Type: "ClusterIP",
		},
	},
}

var ServiceMonitors = map[string]monitoringv1.ServiceMonitor{
	"dedicated-admin-operator": {
		ObjectMeta: metav1.ObjectMeta{
			Name:      operatorconfig.OperatorName,
			Namespace: operatorconfig.OperatorNamespace,
			Labels: map[string]string{
				"k8s-app": operatorconfig.OperatorName,
			},
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Endpoints: []monitoringv1.Endpoint{
				{
					Port:     "metrics",
					Scheme:   "http",
					Interval: "30s",
				},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"k8s-app": operatorconfig.OperatorName,
				},
			},
		},
	},
}
