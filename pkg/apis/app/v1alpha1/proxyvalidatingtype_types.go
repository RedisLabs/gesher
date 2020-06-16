package v1alpha1

import (
	admissionv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ProxyValidatingTypeSpec defines the desired state of ProxyValidatingType
type ProxyValidatingTypeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	Types []admissionv1beta1.RuleWithOperations `json:"types,omitempty" protobuf:"bytes,3,rep,name=types"`
}

// ProxyValidatingTypeStatus defines the observed state of ProxyValidatingType
type ProxyValidatingTypeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ProxyValidatingType is the Schema for the proxyvalidatingtypes API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=proxyvalidatingtypes,scope=Cluster
type ProxyValidatingType struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProxyValidatingTypeSpec   `json:"spec,omitempty"`
	Status ProxyValidatingTypeStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ProxyValidatingTypeList contains a list of ProxyValidatingType
type ProxyValidatingTypeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProxyValidatingType `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ProxyValidatingType{}, &ProxyValidatingTypeList{})
}
