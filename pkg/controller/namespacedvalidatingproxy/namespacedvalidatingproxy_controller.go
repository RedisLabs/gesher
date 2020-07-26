package namespacedvalidatingproxy

import (
	appv1alpha1 "github.com/redislabs/gesher/pkg/apis/app/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_namespacedvalidatingproxy")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new NamespacedValidatingProxy Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileNamespacedValidatingProxy{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("namespacedvalidatingproxy-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource NamespacedValidatingProxy
	err = c.Watch(&source.Kind{Type: &appv1alpha1.NamespacedValidatingProxy{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileNamespacedValidatingProxy implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileNamespacedValidatingProxy{}

// ReconcileNamespacedValidatingProxy reconciles a NamespacedValidatingProxy object
type ReconcileNamespacedValidatingProxy struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a NamespacedValidatingProxy object and makes changes based on the state read
// and what is in the NamespacedValidatingProxy.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileNamespacedValidatingProxy) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling NamespacedValidatingProxy")

	observedState, err := observe(r.client, request, reqLogger)
	if err != nil {
		return reconcile.Result{}, err
	}

	if observedState == nil {
		return reconcile.Result{}, nil
	}

	analyzedState, err := analyze(observedState, reqLogger)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = act(r.client, analyzedState, reqLogger)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}