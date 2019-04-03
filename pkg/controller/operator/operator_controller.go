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

	"github.com/openshift/dedicated-admin-operator/pkg/dedicatedadmin"
	dedicatedadminoperator "github.com/openshift/dedicated-admin-operator/pkg/dedicatedadmin/operator"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("operator-controller")

// Add creates a new Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	k8sClient := mgr.GetClient()
	// Need to pre-load the config ConfigMap to avoid locking/caching issues when
	// 2 or more controllers start simultaneaously and try to load the same object
	dedicatedadmin.GetOperatorConfig(context.Background(), k8sClient)

	return &ReconcileNamespace{client: k8sClient, scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("operator-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Namespace
	err = c.Watch(&source.Kind{Type: &corev1.Namespace{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileNamespace{}

// ReconcileNamespace reconciles a Namespace obj
type ReconcileNamespace struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster Namespace objects and create required operator resources
func (r *ReconcileNamespace) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.Background()
	// Initialize logging obj
	reqLogger := log.WithValues("Request.Namespace", request.Name)

	// Check if the namespace is black listed - administrative namespaces where we
	// don't want to add the dedicated-admin rolebinding, e. g kube-system, openshift-logging
	if request.Name != "openshift-dedicated-admin" {
		reqLogger.Info("Not operator namespace - Skipping")

		return reconcile.Result{}, nil
	}

	// Get the Namespace instance
	ns := &corev1.Namespace{}
	err := r.client.Get(ctx, request.NamespacedName, ns)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, it can be transitioning to the final desired state
			// e. g. deletion or creation still in progress. Return and retry again
			reqLogger.Info("Object not ready")
			return reconcile.Result{}, err
		}
		// Error reading the obj
		reqLogger.Info("Error Getting Namespace")
		return reconcile.Result{}, err
	}

	// Namespace is being deleted
	if ns.Status.Phase == corev1.NamespaceTerminating {
		reqLogger.Info("Namespace Being Deleted")

		return reconcile.Result{}, nil
	}

	// Loop thru our resources and add to the namespace
	for _, obj := range dedicatedadminoperator.ClusterRoleBindings {
		reqLogger.Info("Assigning ClusterRoleBinding to Operator Namespace", "ClusterRoleBinding", obj.Name)

		// Add namespace parameter to resource
		obj.Namespace = request.Name

		err = r.client.Create(ctx, &obj)

		// check for errors, but ignore when resource already exists
		if err != nil && !errors.IsAlreadyExists(err) {
			reqLogger.Info("Error creating ClusterRoleBinding", "ClusterRoleBinding", obj.Name, "Error", err)
			return reconcile.Result{}, err
		}
	}

	for _, obj := range dedicatedadminoperator.Services {
		reqLogger.Info("Assigning Service to Operator Namespace", "Service", obj.Name)

		// Add namespace parameter to resource
		obj.Namespace = request.Name

		err = r.client.Create(ctx, &obj)

		// check for errors, but ignore when resource already exists
		if err != nil && !errors.IsAlreadyExists(err) {
			reqLogger.Info("Error creating Service", "Service", obj.Name, "Error", err)
			return reconcile.Result{}, err
		}
	}

	for _, obj := range dedicatedadminoperator.ServiceMonitors {
		reqLogger.Info("Assigning ServiceMonitor to Operator Namespace", "ServiceMonitor", obj.Name)

		// Add namespace parameter to resource
		obj.Namespace = request.Name

		err = r.client.Create(ctx, &obj)

		// check for errors, but ignore when resource already exists
		if err != nil && !errors.IsAlreadyExists(err) {
			reqLogger.Info("Error creating ServiceMonitor", "ServiceMonitor", obj.Name, "Error", err)
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}
