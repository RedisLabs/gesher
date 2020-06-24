package proxyvalidatingtype

import (
	"github.com/redislabs/gesher/pkg/apis/app/v1alpha1"
	"reflect"

	"github.com/go-logr/logr"
	"k8s.io/api/admissionregistration/v1beta1"
)

type analyzedState struct {
	customResource   *v1alpha1.ProxyValidatingType
	newProxyTypeData *ProxyTypeData
	webhook          *v1beta1.ValidatingWebhookConfiguration
	create           bool
	update           bool
	delete           bool
}

func analyze(observed *observedState, logger logr.Logger) (*analyzedState, error) {
	state := &analyzedState{
		customResource: observed.customResource,
	}

	switch observed.customResource.DeletionTimestamp.IsZero() {
	case true:
		logger.Info("DeletionTimeStamp is zero")
		state.newProxyTypeData = proxyTypeData.Update(observed.customResource)
	case false:
		logger.Info("DeletionTimeStamp is not zero, deleting")
		state.newProxyTypeData = proxyTypeData.Delete(observed.customResource)
		state.delete = true
	}

	webhook := state.newProxyTypeData.GenerateGlobalWebhook()

	// code is ugly to make sure we handle the instance being deleted out from under us
	if webhooksDiffer(webhook, observed.clusterWebhook) {
		logger.Info("Need to update webhook as its changed")
		state.webhook = observed.clusterWebhook
		state.update = true

		if observed.clusterWebhook == nil {
			logger.Info("need to create webhook as it doesn't exist")
			state.webhook = webhook
			state.create = true
		}
		state.webhook.Webhooks = webhook.Webhooks
	}

	return state, nil
}

func webhooksDiffer(new, old *v1beta1.ValidatingWebhookConfiguration) bool {
	if old == nil {
		return true
	}

	return !reflect.DeepEqual(new.Webhooks, old.Webhooks)
}
