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

package admission_proxy

import (
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"os"
)

// toAdmissionResponse is a helper function to create an AdmissionResponse
// with an embedded error
func errToAdmissionResponse(err error) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}

func approved() *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Allowed: true,
	}
}

func GetNamespace() string {
	namespace := os.Getenv("POD_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}
	klog.V(1).Infof("namespace = %v", namespace)

	return namespace
}

func GetServiceName() string {
	serviceName := os.Getenv("PROXY_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "admission"
	}
	klog.V(1).Infof("serviceName = %v", serviceName)

	return serviceName
}
