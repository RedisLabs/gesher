/*
Copyright 2020 Redis Labs Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package namespacedvalidatingtype

import (
	"context"
	"io/ioutil"
	"path/filepath"

	"github.com/redislabs/gesher/pkg/common"
	v1 "k8s.io/api/admissionregistration/v1"

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

var log = logf.Log.WithName("controller_roxyvalidatingtype")

// Add creates a new NamespacedValidatingType Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	var err error

	caBundle, err = ioutil.ReadFile(filepath.Join(common.CertDir, common.CertPem))
	if err != nil {
		return err
	}

	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileNamespacedValidatingType{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("namespacedvalidatingtype-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource NamespacedValidatingType
	err = c.Watch(&source.Kind{Type: &appv1alpha1.NamespacedValidatingType{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &v1.ValidatingWebhookConfiguration{}}, handler.EnqueueRequestsFromMapFunc(
		func(o client.Object) []reconcile.Request {
			if o.GetName() == ProxyWebhookName {
				return []reconcile.Request{{}}
			}
			return nil
		},
	))
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileNamespacedValidatingType implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileNamespacedValidatingType{}

// ReconcileNamespacedValidatingType reconciles a NamespacedValidatingType object
type ReconcileNamespacedValidatingType struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a NamespacedValidatingType object and makes changes based on the state read
// and what is in the NamespacedValidatingType.Spec

// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileNamespacedValidatingType) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling NamespacedValidatingType")

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
