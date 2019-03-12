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

	"github.com/openshift/dedicated-admin-operator/pkg/dedicatedadmin"
	rbacv1 "k8s.io/api/rbac/v1"
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

var log = logf.Log.WithName("rolebinding-controller")

// Add creates a new Rolebinding Controller and adds it to the Manager. The Manager will set fields on the Controller
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

	return &ReconcileRolebinding{client: k8sClient, scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("rolebinding-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes in RoleBindings
	err = c.Watch(&source.Kind{Type: &rbacv1.RoleBinding{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileRolebinding{}

// ReconcileRolebinding reconciles a Rolebinding object
type ReconcileRolebinding struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile watches RoleBindings to
func (r *ReconcileRolebinding) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.Background()

	// Initialize logging object
	reqLogger := log.WithValues("Request.RoleBinding", request.Name)

	// Get operator configuration
	operatorConfig, err := dedicatedadmin.GetOperatorConfig(ctx, r.client)
	if err != nil {
		reqLogger.Info("Error Loading Operator Config", "Error", err)
	}

	// Skip if it's a reserved namespace
	if dedicatedadmin.IsBlackListedNamespace(request.Namespace, operatorConfig.Data["project_blacklist"]) {
		reqLogger.Info("Blacklisted Namespace - Skipping")
		return reconcile.Result{}, nil
	}

	// Fetch the RoleBinding instance
	rb := &rbacv1.RoleBinding{}
	err = r.client.Get(ctx, request.NamespacedName, rb)
	if err != nil {
		if errors.IsNotFound(err) {
			// The RoleBinding was deleted

			// Check if the RB being deleted is from Dedicated Admin
			missingRB, isDedicatedAdminRB := dedicatedadmin.Rolebindings[request.Name]
			if isDedicatedAdminRB {
				reqLogger.Info("Restoring RoleBinding", "Namespace", request.Namespace)

				missingRB.Namespace = request.Namespace
				err = r.client.Create(ctx, &missingRB)
				if err != nil {
					reqLogger.Info("Error creating rolebinding", "RoleBinding", missingRB.Name, "Error", err)
					return reconcile.Result{}, err
				}
			}

			return reconcile.Result{}, nil
		}
		// Error reading the object
		reqLogger.Info("Error Getting RoleBinding")
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
