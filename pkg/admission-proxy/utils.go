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
