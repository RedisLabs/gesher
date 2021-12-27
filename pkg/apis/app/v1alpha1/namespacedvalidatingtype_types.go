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
	admissionv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NamespacedValidatingTypeSpec defines the desired state of NamespacedValidatingType
type NamespacedValidatingTypeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	Types []admissionv1.RuleWithOperations `json:"types,omitempty" protobuf:"bytes,3,rep,name=types"`
}

// NamespacedValidatingTypeStatus defines the observed state of NamespacedValidatingType
type NamespacedValidatingTypeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NamespacedValidatingType is the Schema for the namespacedvalidatingtypes API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=namespacedvalidatingtype,scope=Cluster
type NamespacedValidatingType struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NamespacedValidatingTypeSpec   `json:"spec,omitempty"`
	Status NamespacedValidatingTypeStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NamespacedValidatingTypeList contains a list of NamespacedValidatingType
type NamespacedValidatingTypeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NamespacedValidatingType `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NamespacedValidatingType{}, &NamespacedValidatingTypeList{})
}

func (pvt *NamespacedValidatingType) GetObservedGeneration() int64 {
	return pvt.Status.ObservedGeneration
}
