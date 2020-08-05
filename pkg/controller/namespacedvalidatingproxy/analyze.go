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
		logger.V(2).Info("DeletionTimeStamp is zero")
		state.newEndpointData = EndpointData.Update(observed.customResource)
	case false:
		logger.V(2).Info("DeletionTimeStamp is not zero, deleting")
		state.newEndpointData = EndpointData.Delete(observed.customResource)
		state.delete = true
	}

	if !reflect.DeepEqual(state.newEndpointData, EndpointData) {
		state.update = true
	}

	return state, nil
}
