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
package rolebinding

import (
	"context"
	"testing"

	"github.com/openshift/dedicated-admin-operator/config"
	dedicatedadminproject "github.com/openshift/dedicated-admin-operator/pkg/dedicatedadmin/project"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	//	realclient "sigs.k8s.io/controller-runtime/pkg/client"
	client "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func reset(ctx context.Context, r *ReconcileRolebinding) {
	r.client.Delete(ctx, makeConfig())
	for _, rb := range dedicatedadminproject.RoleBindings {
		r.client.Delete(ctx, rb.DeepCopyObject())
	}
	r.client.Delete(ctx, makeNamespace("test"))
}

func makeTestReconciler() *ReconcileRolebinding {
	return &ReconcileRolebinding{
		client: client.NewFakeClient(),
		scheme: nil,
	}
}

func TestReconcileIgnoresForeignRoleBinding(t *testing.T) {
	// Don't try to re-create RoleBindings that we don't explicitly care about.
	ctx := context.TODO()
	reconciler := makeTestReconciler()
	defer reset(ctx, reconciler)

	cerr := reconciler.client.Create(ctx, makeConfig())
	if cerr != nil {
		t.Error("Couldn't create the required configmap for the test")
	}
	nerr := reconciler.client.Create(ctx, makeNamespace("test"))
	if nerr != nil {
		t.Error("Couldn't create the required namespace for the test")
	}

	request := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-foreign-rb",
			Namespace: "test",
		},
	}
	res, err := reconciler.Reconcile(request)
	if res.Requeue {
		t.Error("Didn't expect to requeue")
	}
	if err != nil {
		t.Errorf("Got an unexpected error: %s", err)
	}

	remadeRB := rbacv1.RoleBinding{}
	err = reconciler.client.Get(ctx, request.NamespacedName, &remadeRB)
	if err == nil {
		t.Errorf("Didn't expected the rolebinding %s to be remade, and it was:  %s", request.NamespacedName, err)
	}

}

func TestCreatesRoleBindingInCorrectNS(t *testing.T) {
	// should create a RoleBinding in a namespace named "test"
	// No need to pre-create the RoleBinding because when Reconcile is called, it
	// will be treated as though it was deleted, so pre-absent suits the requirement.
	ctx := context.TODO()
	reconciler := makeTestReconciler()
	defer reset(ctx, reconciler)

	cerr := reconciler.client.Create(ctx, makeConfig())
	if cerr != nil {
		t.Error("Couldn't create the required configmap for the test")
	}
	nerr := reconciler.client.Create(ctx, makeNamespace("test"))
	if nerr != nil {
		t.Error("Couldn't create the required namespace for the test")
	}
	rbname := ""
	// just need one key
	for rbn := range dedicatedadminproject.RoleBindings {
		rbname = rbn
		break
	}

	request := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      rbname,
			Namespace: "test",
		},
	}

	res, err := reconciler.Reconcile(request)
	if res.Requeue {
		t.Error("Didn't expect to requeue")
	}
	if err != nil {
		t.Errorf("Got an unexpected error: %s", err)
	}

	remadeRB := rbacv1.RoleBinding{}
	err = reconciler.client.Get(ctx, request.NamespacedName, &remadeRB)
	if err != nil {
		t.Errorf("Expected the rolebinding %s to be remade, and it wasn't:  %s", request.NamespacedName, err)
	}
}

func TestBlockedNamespace(t *testing.T) {
	// No RoleBinding should be recreated
	ctx := context.TODO()
	reconciler := makeTestReconciler()
	defer reset(ctx, reconciler)

	cerr := reconciler.client.Create(ctx, makeConfig())
	if cerr != nil {
		t.Error("Couldn't create the required configmap for the test")
	}

	request := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-rb",
			Namespace: "logging",
		},
	}

	res, err := reconciler.Reconcile(request)
	if res.Requeue {
		t.Error("Didn't expect to requeue")
	}
	if err != nil {
		t.Errorf("Got an unexpected error: %s", err)
	}

	remadeRB := rbacv1.RoleBinding{}
	err = reconciler.client.Get(ctx, request.NamespacedName, &remadeRB)
	if err == nil {
		t.Errorf("Didn't expected the rolebinding %s to be remade, and it was:  %s", request.NamespacedName, err)
	}

}

func makeNamespace(ns string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: ns},
	}
}

func makeConfig() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.OperatorConfigMapName,
			Namespace: config.OperatorNamespace,
		},
		Data: map[string]string{
			"project_blacklist": "^kube-.*,^openshift-.*,^logging$,^default$,^openshift$",
		},
	}
}
