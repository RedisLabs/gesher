package namespacedvalidatingproxy

import (
	"github.com/go-logr/logr"
	"github.com/redislabs/gesher/pkg/apis/app/v1alpha1"
	"reflect"
)

type analyzedState struct {
	customResource *v1alpha1.NamespacedValidatingProxy
	newEndpointData *EndpointDataType
	update bool
	delete bool
}

func analyze(observed *observeState, logger logr.Logger) (*analyzedState, error) {
	state := &analyzedState{
		customResource: observed.customResource,
	}

	switch observed.customResource.DeletionTimestamp.IsZero() {
	case true:
		logger.Info("DeletionTimeStamp is zero")
		state.newEndpointData = EndpointData.Update(observed.customResource)
	case false:
		logger.Info("DeletionTimeStamp is not zero, deleting")
		state.newEndpointData = EndpointData.Delete(observed.customResource)
		state.delete = true
	}

	if !reflect.DeepEqual(state.newEndpointData, EndpointData) {
		state.update = true
	}

	return state, nil
}
