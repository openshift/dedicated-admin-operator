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
