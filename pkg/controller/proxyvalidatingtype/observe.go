package proxyvalidatingtype

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"

	appv1alpha1 "github.com/redislabs/gesher/pkg/apis/app/v1alpha1"
	"k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type observedState struct {
	customResource *appv1alpha1.ProxyValidatingType
	clusterWebhook *v1beta1.ValidatingWebhookConfiguration
}

func observe(client client.Client, request reconcile.Request, logger logr.Logger) (*observedState, error) {
	state := &observedState{
		customResource: &appv1alpha1.ProxyValidatingType{},
		clusterWebhook: &v1beta1.ValidatingWebhookConfiguration{},
	}

	// Fetch the ProxyValidatingType instance
	if request.Name != "" {
		err := client.Get(context.TODO(), request.NamespacedName, state.customResource)
		if err != nil {
			if errors.IsNotFound(err) {
				logger.Info("didn't find resource")
				// Request object not found, could have been deleted after reconcile request.
				// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
				// Return and don't requeue
				return nil, nil
			}

			// Error reading the object
			logger.Error(err, "resource retrieval failed")
			return nil, err
		}
	} else {
		state.customResource = nil
	}

	// Fetch the managed ValidatingWebhookConfiguration instance
	// code is ugly to make sure we handle the instance being deleted out from under us
	err := client.Get(context.TODO(), types.NamespacedName{Name: ProxyWebhookName}, state.clusterWebhook)
	if err != nil {
		if !errors.IsNotFound(err) {
			// Error reading the object
			return nil, err
		}
		logger.V(2).Info("cluster webhook doesn't exist yet")
		state.clusterWebhook = nil
	} else {
		logger.V(2).Info(fmt.Sprintf("clusterWebhook = %+v", state.clusterWebhook))
	}

	return state, nil
}
