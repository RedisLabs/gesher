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
	admregv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

// NamespacedValidatingRuleSpec defines the desired state of NamespacedValidatingRule
type NamespacedValidatingRuleSpec struct {
	// Webhooks is a list of webhooks and the affected resources and operations.
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Webhooks []admregv1.ValidatingWebhook `json:"webhooks,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,2,rep,name=Webhooks"`
}

// NamespacedValidatingRuleStatus defines the observed state of NamespacedValidatingRule
type NamespacedValidatingRuleStatus struct {
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NamespacedValidatingRule is the Schema for the namespacedvalidatingrule API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=namespacedvalidatingrule,scope=Namespaced
type NamespacedValidatingRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NamespacedValidatingRuleSpec   `json:"spec,omitempty"`
	Status NamespacedValidatingRuleStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NamespacedValidatingRuleyList contains a list of NamespacedValidatingRule
type NamespacedValidatingRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NamespacedValidatingRule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NamespacedValidatingRule{}, &NamespacedValidatingRuleList{})
}

func (nvp *NamespacedValidatingRule) GetObservedGeneration() int64 {
	return nvp.Status.ObservedGeneration
}
