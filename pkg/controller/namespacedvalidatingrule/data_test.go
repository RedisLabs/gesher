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

package namespacedvalidatingrule

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/redislabs/gesher/pkg/apis/app/v1alpha1"
)

const (
	uid1 = "1"
	uid2 = "2"

	namespace = "test"

	testGroup1    = "testGroup1"
	testVersion1  = "testVersion1"
	testResource1 = "testResource1"
	testOp1       = v1beta1.Create
	testOp2       = v1beta1.Delete
)

var (
	resource1 = &v1alpha1.NamespacedValidatingRule{
		ObjectMeta: metav1.ObjectMeta{
			UID:       uid1,
			Namespace: namespace,
		},
		Spec: v1alpha1.NamespacedValidatingRuleSpec{
			Webhooks: []v1beta1.ValidatingWebhook{{
				Name:         "resource1",
				ClientConfig: v1beta1.WebhookClientConfig{},
				Rules: []v1beta1.RuleWithOperations{{
					Operations: []v1beta1.OperationType{testOp1},
					Rule: v1beta1.Rule{
						APIGroups:   []string{testGroup1},
						APIVersions: []string{testVersion1},
						Resources:   []string{testResource1},
					},
				}},
			}},
		},
	}

	resource1a = &v1alpha1.NamespacedValidatingRule{
		ObjectMeta: metav1.ObjectMeta{
			UID:       uid1,
			Namespace: namespace,
		},
		Spec: v1alpha1.NamespacedValidatingRuleSpec{
			Webhooks: []v1beta1.ValidatingWebhook{{
				Name:         "resource1",
				ClientConfig: v1beta1.WebhookClientConfig{},
				Rules: []v1beta1.RuleWithOperations{{
					Operations: []v1beta1.OperationType{testOp2},
					Rule: v1beta1.Rule{
						APIGroups:   []string{testGroup1},
						APIVersions: []string{testVersion1},
						Resources:   []string{testResource1},
					},
				}},
			}},
		},
	}

	resource2 = &v1alpha1.NamespacedValidatingRule{
		ObjectMeta: metav1.ObjectMeta{
			UID:       uid2,
			Namespace: namespace,
		},
		Spec: v1alpha1.NamespacedValidatingRuleSpec{
			Webhooks: []v1beta1.ValidatingWebhook{{
				Name: "resource2",
				ClientConfig: v1beta1.WebhookClientConfig{
					Service:  &v1beta1.ServiceReference{Namespace: namespace},
					CABundle: nil,
				},
				Rules: []v1beta1.RuleWithOperations{{
					Operations: []v1beta1.OperationType{testOp1},
					Rule: v1beta1.Rule{
						APIGroups:   []string{testGroup1},
						APIVersions: []string{testVersion1},
						Resources:   []string{testResource1},
					},
				}},
			}},
		},
	}

	resource3 = &v1alpha1.NamespacedValidatingRule{
		ObjectMeta: metav1.ObjectMeta{
			UID:       uid2,
			Namespace: namespace,
		},
		Spec: v1alpha1.NamespacedValidatingRuleSpec{
			Webhooks: []v1beta1.ValidatingWebhook{{
				Name: "resource2",
				ClientConfig: v1beta1.WebhookClientConfig{
					Service:  &v1beta1.ServiceReference{},
					CABundle: nil,
				},
				Rules: []v1beta1.RuleWithOperations{{
					Operations: []v1beta1.OperationType{testOp1},
					Rule: v1beta1.Rule{
						APIGroups:   []string{testGroup1},
						APIVersions: []string{testVersion1},
						Resources:   []string{testResource1},
					},
				}},
			}},
		},
	}

)

func TestAdd(t *testing.T) {
	endpoindData := &EndpointDataType{}
	newE := endpoindData.Add(resource1)

	groupMap, ok := newE.Mapping[namespace]
	assert.True(t, ok)
	assert.NotEmpty(t, groupMap)

	versionMap, ok := groupMap[testGroup1]
	assert.True(t, ok)
	assert.NotEmpty(t, versionMap)

	resourceMap, ok := versionMap[testVersion1]
	assert.True(t, ok)
	assert.NotEmpty(t, resourceMap)

	opMap, ok := resourceMap[testResource1]
	assert.True(t, ok)
	assert.NotEmpty(t, opMap)

	instanceMap, ok := opMap[testOp1]
	assert.True(t, ok)
	assert.NotEmpty(t, instanceMap)
	_, ok = instanceMap[uid1]
	assert.True(t, ok)
}

func TestDelete(t *testing.T) {
	endpoindData := &EndpointDataType{}
	newE := endpoindData.Add(resource1)
	newE = newE.Delete(resource1)

	groupMap, ok := newE.Mapping[namespace]
	assert.True(t, ok)
	assert.NotEmpty(t, groupMap)

	versionMap, ok := groupMap[testGroup1]
	assert.True(t, ok)
	assert.NotEmpty(t, versionMap)

	resourceMap, ok := versionMap[testVersion1]
	assert.True(t, ok)
	assert.NotEmpty(t, resourceMap)

	opMap, ok := resourceMap[testResource1]
	assert.True(t, ok)
	assert.NotEmpty(t, opMap)

	instanceMap, ok := opMap[testOp1]
	assert.True(t, ok)
	assert.Empty(t, instanceMap)
}

func TestUpdate(t *testing.T) {
	endpoindData := &EndpointDataType{}
	newE := endpoindData.Add(resource1)
	newE = newE.Update(resource1a)

	groupMap, ok := newE.Mapping[namespace]
	assert.True(t, ok)
	assert.NotEmpty(t, groupMap)

	versionMap, ok := groupMap[testGroup1]
	assert.True(t, ok)
	assert.NotEmpty(t, versionMap)

	resourceMap, ok := versionMap[testVersion1]
	assert.True(t, ok)
	assert.NotEmpty(t, resourceMap)

	opMap, ok := resourceMap[testResource1]
	assert.True(t, ok)
	assert.NotEmpty(t, opMap)

	instanceMap, ok := opMap[testOp1]
	assert.True(t, ok)
	assert.Empty(t, instanceMap)

	instanceMap, ok = opMap[testOp2]
	assert.True(t, ok)
	assert.NotEmpty(t, instanceMap)
	_, ok = instanceMap[uid1]
	assert.True(t, ok)
}

func TestGet(t *testing.T) {
	endpoindData := &EndpointDataType{}
	newE := endpoindData.Add(resource2)
	w := newE.Get(namespace, metav1.GroupVersionResource{Group: testGroup1, Version: testVersion1, Resource: testResource1}, testOp1)
	assert.NotEmpty(t, w)
	assert.Len(t, w, 1)
	assert.Equal(t, w[0].ClientConfig.Service.Namespace, namespace)
}

func TestRuleNoNamespace(t *testing.T) {
	endpoindData := &EndpointDataType{}
	assert.Equal(t, "", resource3.Spec.Webhooks[0].ClientConfig.Service.Namespace, "resource3 doesn''t have an empty service namespace")
	newE := endpoindData.Add(resource3)
	w := newE.Get(namespace, metav1.GroupVersionResource{Group: testGroup1, Version: testVersion1, Resource: testResource1}, testOp1)
	assert.NotEmpty(t, w)
	assert.Len(t, w, 1)
	assert.Equal(t, w[0].ClientConfig.Service.Namespace, namespace)
}