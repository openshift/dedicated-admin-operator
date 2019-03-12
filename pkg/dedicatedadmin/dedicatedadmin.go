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

	"github.com/openshift/dedicated-admin-operator/config"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
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
	var configMap types.NamespacedName

	configMap.Name = config.OperatorConfigMapName
	configMap.Namespace = config.OperatorNamespace
	// daLogger.Info("GetOperatorConfig", "ConfigMap Get Request", configMap)

	// Load config map with operator's config
	operatorConfig := &corev1.ConfigMap{}
	err := k8sClient.Get(ctx, configMap, operatorConfig)

	return operatorConfig, err
}
