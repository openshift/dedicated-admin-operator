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

package namespace

import (
	"context"

	"github.com/rogbas/dedicated-admin-operator/pkg/dedicatedadmin"
	"github.com/rogbas/dedicated-admin-operator/pkg/metrics"
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

var log = logf.Log.WithName("namespace-controller")

// Add creates a new Project Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileNamespace{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("dedicated-admin-namespace-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	metrics.StartMetrics()

	// Watch for changes to primary resource Project
	err = c.Watch(&source.Kind{Type: &corev1.Namespace{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileNamespace{}

// ReconcileNamespace reconciles a Project object
type ReconcileNamespace struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster Namespace objects and assign proper
// rolebindings when applicable (not black listed)
func (r *ReconcileNamespace) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.Background()

	return reconcile.Result{}, nil
	// Initialize logging object
	reqLogger := log.WithValues("Request.Namespace", request.Name)

	// Get operator configuration
	operatorConfig, err := dedicatedadmin.GetOperatorConfig(ctx, r.client)
	if err != nil {
		reqLogger.Info("Error Loading Operator Config", "Error", err)
	}

	// reqLogger.Info("Loaded operator config", "config", operatorConfig.Data)

	// Check if the namespace is black listed - administrative namespaces where we
	// don't want to add the dedicated-admin rolebinding, e. g kube-system, openshift-logging
	if dedicatedadmin.IsBlackListedNamespace(request.Name, operatorConfig.Data["project_blacklist"]) {
		reqLogger.Info("Blacklisted Namespace - Skipping")

		// Increment counter on prometheus
		metrics.IncBlacklistedCount()

		return reconcile.Result{}, nil
	}

	// Get the Namespace instance
	ns := &corev1.Namespace{}
	err = r.client.Get(ctx, request.NamespacedName, ns)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, it can be transitioning to the final desired state
			// e. g. deletion or creation still in progress. Return and retry again
			reqLogger.Info("Object not ready")
			return reconcile.Result{}, nil
		}
		// Error reading the object
		reqLogger.Info("Error Getting Namespace")
		return reconcile.Result{}, err
	}
	// Namespace is being deleted, return and retry the reconcile loop
	if ns.Status.Phase == corev1.NamespaceTerminating {
		reqLogger.Info("Namespace Being Deleted")
		return reconcile.Result{}, nil
	}

	// Loop thru our map of rolebindings, adding each one to the namespace
	for _, rb := range dedicatedadmin.Rolebindings {
		reqLogger.Info("Assigning RoleBinding to Namespace", "RoleBinding", rb.Name)

		// Add namespace parameter to rolebinding
		rb.Namespace = request.Name

		err = r.client.Create(ctx, &rb)
		// check for errors, but ignore when rolebinding already exists
		if err != nil && !errors.IsAlreadyExists(err) {
			reqLogger.Info("Error creating rolebinding", "RoleBinding", rb.Name, "Error", err)
		}
	}

	return reconcile.Result{}, nil
}
