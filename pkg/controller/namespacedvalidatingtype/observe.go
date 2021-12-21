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
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"

	appv1alpha1 "github.com/redislabs/gesher/pkg/apis/app/v1alpha1"
	admregv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type observedState struct {
	customResource *appv1alpha1.NamespacedValidatingType
	clusterWebhook *admregv1.ValidatingWebhookConfiguration
}

func observe(client client.Client, request reconcile.Request, logger logr.Logger) (*observedState, error) {
	state := &observedState{
		customResource: &appv1alpha1.NamespacedValidatingType{},
		clusterWebhook: &admregv1.ValidatingWebhookConfiguration{},
	}

	// Fetch the NamespacedValidatingType instance
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
