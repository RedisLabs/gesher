package namespacedvalidatingproxy

import (
	"context"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	proxyFinalizer = "proxy.finalizer.gesher"
)


func act(kubeClient client.Client, state *analyzedState, logger logr.Logger) error {
	var fullChange bool
	ret := manageFinalizer(state, logger)
	fullChange = ret || fullChange

	var statusChange bool
	ret = manageGeneration(state, logger)
	statusChange = ret || statusChange

	if fullChange {
		logger.Info("doing full update")
		err := kubeClient.Update(context.TODO(), state.customResource)
		if err != nil {
			logger.Error(err, "failed to do full update")
			return err
		}
	} else if statusChange {
		logger.Info("doing status update")
		err := kubeClient.Status().Update(context.TODO(), state.customResource)
		if err != nil {
			logger.Error(err, "failed to do status update")
			return err
		}
	}

	EndpointData = state.newEndpointData

	return nil
}

func manageFinalizer(state *analyzedState, logger logr.Logger) bool {
	var ret bool

	switch state.delete {
	case false:
		if !containsString(state.customResource.ObjectMeta.Finalizers, proxyFinalizer) {
			logger.Info("adding finalizer")
			state.customResource.ObjectMeta.Finalizers = append(state.customResource.ObjectMeta.Finalizers, proxyFinalizer)
			ret = true
		}
	case true:
		if containsString(state.customResource.ObjectMeta.Finalizers, proxyFinalizer) {
			logger.Info("removing finalizer")
			state.customResource.ObjectMeta.Finalizers = removeString(state.customResource.ObjectMeta.Finalizers, proxyFinalizer)
			ret = true
		}
	}

	return ret
}

func manageGeneration(state *analyzedState, logger logr.Logger) bool {
	var ret bool

	if state.customResource.Status.ObservedGeneration != state.customResource.Generation {
		logger.Info("updating observed generation in status")
		state.customResource.Status.ObservedGeneration = state.customResource.Generation
		ret = true
	}

	return ret
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}