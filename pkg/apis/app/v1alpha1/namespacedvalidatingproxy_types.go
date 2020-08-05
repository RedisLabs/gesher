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

package v1alpha1

import (
	"k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html


// NamespacedValidatingProxySpec defines the desired state of NamespacedValidatingProxy
type NamespacedValidatingProxySpec struct {
	// Webhooks is a list of webhooks and the affected resources and operations.
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Webhooks []v1beta1.ValidatingWebhook `json:"webhooks,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,2,rep,name=Webhooks"`
}

// NamespacedValidatingProxyStatus defines the observed state of NamespacedValidatingProxy
type NamespacedValidatingProxyStatus struct {
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NamespacedValidatingProxy is the Schema for the namespacedvalidatingproxy API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=namespacedvalidatingproxy,scope=Namespaced
type NamespacedValidatingProxy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NamespacedValidatingProxySpec   `json:"spec,omitempty"`
	Status NamespacedValidatingProxyStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NamespacedValidatingProxyList contains a list of NamespacedValidatingProxy
type NamespacedValidatingProxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NamespacedValidatingProxy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NamespacedValidatingProxy{}, &NamespacedValidatingProxyList{})
}

func (nvp *NamespacedValidatingProxy) GetObservedGeneration() int64 {
	return nvp.Status.ObservedGeneration
}