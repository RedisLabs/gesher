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
	"reflect"

	"github.com/redislabs/gesher/pkg/apis/app/v1alpha1"

	"github.com/go-logr/logr"
	admregv1 "k8s.io/api/admissionregistration/v1"
)

type analyzedState struct {
	customResource        *v1alpha1.NamespacedValidatingType
	newNamespacedTypeData *NamespacedTypeData
	webhook               *admregv1.ValidatingWebhookConfiguration
	create                bool
	update                bool
	delete                bool
}

func analyze(observed *observedState, logger logr.Logger) (*analyzedState, error) {
	state := &analyzedState{
		customResource:        observed.customResource,
		newNamespacedTypeData: namespacedTypeData,
	}

	if state.customResource != nil {
		switch observed.customResource.DeletionTimestamp.IsZero() {
		case true:
			logger.V(2).Info("DeletionTimeStamp is zero")
			state.newNamespacedTypeData = namespacedTypeData.Update(observed.customResource)
		case false:
			logger.V(2).Info("DeletionTimeStamp is not zero, deleting")
			state.newNamespacedTypeData = namespacedTypeData.Delete(observed.customResource)
			state.delete = true
		}
	}

	webhook := state.newNamespacedTypeData.GenerateGlobalWebhook()

	// code is ugly to make sure we handle the instance being deleted out from under us
	if webhooksDiffer(webhook, observed.clusterWebhook) {
		logger.V(2).Info("Need to update webhook as its changed")
		state.webhook = observed.clusterWebhook
		state.update = true

		if state.webhook == nil {
			logger.V(2).Info("need to create webhook as it doesn't exist")
			state.webhook = webhook
			state.create = true
		}
		state.webhook.Webhooks = webhook.Webhooks
	}

	return state, nil
}

func webhooksDiffer(new, old *admregv1.ValidatingWebhookConfiguration) bool {
	if old == nil {
		return true
	}

	return !reflect.DeepEqual(new.Webhooks, old.Webhooks)
}
