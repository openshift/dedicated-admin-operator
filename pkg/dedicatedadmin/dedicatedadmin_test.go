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
	"testing"
)

// TestBlacklist exercises the comma-separating in IsBlackListedNamespace
func TestBlacklist(t *testing.T) {
	var tests = []struct {
		configmapstring string // what we might expect from the ConfigMap
		challenge       string // what would the test namespace be against the regexp?
		valid           bool   // would we expect it to be a real regexp?
	}{
		{"^kube-.*,^openshift-.*,^logging$,^default$,^openshift$", "openshift-test", true},
		{"", "", true},
		{"nonexpr", "openshift-test", false},
		{"openshift", "openshift-test", true},
		{"openshift,kube", "kube-system", true},
		{"^kube-(system|default|public)", "kube-bar", false},
		{"^kube-(system|default|public)", "kube-system", true},
		{"^kube-(system|default|public)", "kube-default-baz", true},
		{"^(kube-(system|default|foo)|openshift-.*).*$", "kube-default-baz", true},
		{"^(kube-(system|default|foo)|openshift-.*).*$", "kube-baz", false},
	}
	for _, test := range tests {
		if IsBlackListedNamespace(test.challenge, test.configmapstring) != test.valid {
			t.Errorf("challenge `%s` against regex str `%s` not %t (got %t)", test.challenge, test.configmapstring, test.valid, !test.valid)
		}
	}
}
