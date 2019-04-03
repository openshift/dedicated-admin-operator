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
	"context"
	"testing"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	dedicatedadminoperator "github.com/openshift/dedicated-admin-operator/pkg/dedicatedadmin/operator"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	realclient "sigs.k8s.io/controller-runtime/pkg/client"
	client "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var nsName = "openshift-dedicated-admin"

func reset(ctx context.Context, r *ReconcileNamespace) {
	r.client.Delete(ctx, makeNamespace(nsName))
}

func makeTestReconciler() *ReconcileNamespace {
	return &ReconcileNamespace{
		client: client.NewFakeClient(),
		scheme: nil,
	}
}

func TestInvalidNamespace(t *testing.T) {
	ctx := context.TODO()
	reconciler := makeTestReconciler()
	defer reset(ctx, reconciler)

	request := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "someothernamespace",
			Namespace: "",
		},
	}

	res, err := reconciler.Reconcile(request)
	if res.Requeue {
		t.Error("Expected not to requeue, but we were")
	}
	if err != nil {
		t.Errorf("Expected no error, but got one: %s", err)
	}
}

func TestValidNamespace(t *testing.T) {
	monitoringv1.AddToScheme(scheme.Scheme)

	ctx := context.TODO()
	reconciler := makeTestReconciler()
	defer reset(ctx, reconciler)

	// create namespace needed for the test
	nerr := reconciler.client.Create(ctx, makeNamespace(nsName))
	if nerr != nil {
		t.Errorf("Couldn't create the required namespace: %s", nsName)
	}

	request := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      nsName,
			Namespace: "",
		},
	}

	res, err := reconciler.Reconcile(request)
	if res.Requeue {
		t.Error("Expected not to requeue, but we did")
	}
	if err != nil {
		t.Errorf("Expected no error, but got one: %s", err)
	}

	// check resources: clusterrolebindings, services, servicemonitors

	// clusterrolebindings:
	{
		list := rbacv1.ClusterRoleBindingList{}
		opts := realclient.ListOptions{Namespace: request.Name}

		err = reconciler.client.List(ctx, &opts, &list)
		if err != nil {
			t.Errorf("Error listing %s: %s", "ClusterRoleBindings", err)
		}

		seen := make(map[string]bool)
		for _, obj := range dedicatedadminoperator.ClusterRoleBindings {
			seen[obj.ObjectMeta.Name] = false
		}
		for _, obj := range list.Items {
			seen[obj.ObjectMeta.Name] = true
		}
		for obj_name, s := range seen {
			if !s {
				t.Errorf("Expected but didn't see %s: %s", "ClusterRoleBinding", obj_name)
			}
		}
	}

	// services:
	{
		list := corev1.ServiceList{}
		opts := realclient.ListOptions{Namespace: request.Name}

		err = reconciler.client.List(ctx, &opts, &list)
		if err != nil {
			t.Errorf("Error listing %s: %s", "Serivces", err)
		}

		seen := make(map[string]bool)
		for _, obj := range dedicatedadminoperator.Services {
			seen[obj.ObjectMeta.Name] = false
		}
		for _, obj := range list.Items {
			seen[obj.ObjectMeta.Name] = true
		}
		for obj_name, s := range seen {
			if !s {
				t.Errorf("Expected but didn't see %s: %s", "Service", obj_name)
			}
		}
	}

	// servicemonitors:
	{
		list := monitoringv1.ServiceMonitorList{}
		opts := realclient.ListOptions{Namespace: request.Name}

		err = reconciler.client.List(ctx, &opts, &list)
		if err != nil {
			t.Errorf("Error listing %s: %s", "SerivceMonitors", err)
		}

		seen := make(map[string]bool)
		for _, obj := range dedicatedadminoperator.ServiceMonitors {
			seen[obj.ObjectMeta.Name] = false
		}
		for _, obj := range list.Items {
			seen[obj.ObjectMeta.Name] = true
		}
		for obj_name, s := range seen {
			if !s {
				t.Errorf("Expected but didn't see %s: %s", "SerivceMonitors", obj_name)
			}
		}
	}
}

func makeNamespace(ns string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ns,
			Namespace: "",
		},
	}
}
