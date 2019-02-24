package rolebinding

import (
	"context"

	"github.com/rogbas/dedicated-admin-operator/pkg/dedicatedadmin"
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
	return &ReconcileRolebinding{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("rolebinding-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Rolebinding
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

// Reconcile reads that state of the cluster for a Rolebinding object and makes changes based on the state read
// and what is in the Rolebinding.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
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

	reqLogger.Info("Reconciling Rolebinding")

	// Fetch the RoleBinding instance
	rb := &rbacv1.RoleBinding{}
	err = r.client.Get(ctx, request.NamespacedName, rb)
	if err != nil {
		if errors.IsNotFound(err) {
			// The RoleBinding was deleted

			// Check if the RB being deleted is from Dedicated Admin
			missingRB, isDedicatedAdminRB := dedicatedadmin.Rolebindings[request.Name]
			if isDedicatedAdminRB {
				reqLogger.Info("Restoring RoleBinding")

				missingRB.Namespace = request.Namespace
				err = r.client.Create(ctx, &missingRB)
				if err != nil {
					reqLogger.Info("Error creating rolebinding", "RoleBinding", missingRB.Name, "Error", err)
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
