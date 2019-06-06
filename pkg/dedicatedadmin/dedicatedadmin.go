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

package dedicatedadmin

import (
	"context"
	"regexp"
	"strings"

	operatorconfig "github.com/openshift/dedicated-admin-operator/config"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var (
	log      = logf.Log.WithName("dedicatedadmin")
	daLogger = log.WithValues("DedicatedAdmin", "functions")
)

// IsBlackListedNamespace matchs a nam,espace against the blacklist
func IsBlackListedNamespace(namespace string, blacklistedNamespaces string) bool {
	for _, blackListedNS := range strings.Split(blacklistedNamespaces, ",") {
		matched, _ := regexp.MatchString(blackListedNS, namespace)
		if matched {
			return true
		}
	}
	return false
}

// GetOperatorConfig gets the operator's configuration from a config map
func GetOperatorConfig(ctx context.Context, k8sClient client.Client) (*corev1.ConfigMap, error) {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      operatorconfig.OperatorConfigMapName,
			Namespace: operatorconfig.OperatorNamespace,
		},
		// Always update the PrometheusRule when updating this regexp, reflecting the same changes on the expr for the alert rule
		Data: map[string]string{
			"project_blacklist": "^kube-.*,^openshift-.*,^logging$,^default$,^openshift$,^ops-health-monitoring$,^ops-project-operation-check$,^management-infra$",
		},
	}, nil
}
