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

package namespacedvalidatingtype

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/redislabs/gesher/pkg/apis/app/v1alpha1"
)

const (
	uid1 = "1"
	uid2 = "2"
	uid3 = "3"

	testGroup1   = "testGroup1"
	testVersion1 = "testVersion1"
	testKind1    = "testKind1"
	testGroup2   = "testGroup2"
	testVersion2 = "testVersion2"
	testKind2    = "testKind2"
	testOp1      = v1beta1.Create
	testOp2      = v1beta1.Delete
)

var (
	resource1 = &v1alpha1.NamespacedValidatingType{
		ObjectMeta: metav1.ObjectMeta{UID: uid1},
		Spec: v1alpha1.NamespacedValidatingTypeSpec{
			Types: []v1beta1.RuleWithOperations{{
				Operations: []v1beta1.OperationType{testOp1},
				Rule: v1beta1.Rule{
					APIGroups:   []string{testGroup1},
					APIVersions: []string{testVersion1},
					Resources:   []string{testKind1},
				},
			}},
		},
	}
	resource1a = &v1alpha1.NamespacedValidatingType{
		ObjectMeta: metav1.ObjectMeta{UID: uid1},
		Spec: v1alpha1.NamespacedValidatingTypeSpec{
			Types: []v1beta1.RuleWithOperations{{
				Operations: []v1beta1.OperationType{testOp2},
				Rule: v1beta1.Rule{
					APIGroups:   []string{testGroup1},
					APIVersions: []string{testVersion1},
					Resources:   []string{testKind1},
				},
			}},
		},
	}
	resource2 = &v1alpha1.NamespacedValidatingType{
		ObjectMeta: metav1.ObjectMeta{UID: uid2},
		Spec: v1alpha1.NamespacedValidatingTypeSpec{
			Types: []v1beta1.RuleWithOperations{{
				Operations: []v1beta1.OperationType{testOp1},
				Rule: v1beta1.Rule{
					APIGroups:   []string{testGroup1},
					APIVersions: []string{testVersion1},
					Resources:   []string{testKind1},
				},
			}},
		},
	}
	resource2a = &v1alpha1.NamespacedValidatingType{
		ObjectMeta: metav1.ObjectMeta{UID: uid2},
		Spec: v1alpha1.NamespacedValidatingTypeSpec{
			Types: []v1beta1.RuleWithOperations{{
				Operations: []v1beta1.OperationType{testOp2},
				Rule: v1beta1.Rule{
					APIGroups:   []string{testGroup1},
					APIVersions: []string{testVersion1},
					Resources:   []string{testKind1},
				},
			}},
		},
	}
	resource3 = &v1alpha1.NamespacedValidatingType{
		ObjectMeta: metav1.ObjectMeta{UID: uid3},
		Spec: v1alpha1.NamespacedValidatingTypeSpec{
			Types: []v1beta1.RuleWithOperations{{
				Operations: []v1beta1.OperationType{testOp1},
				Rule: v1beta1.Rule{
					APIGroups:   []string{testGroup2},
					APIVersions: []string{testVersion2},
					Resources:   []string{testKind2},
				},
			}},
		},
	}
)

func TestAdd(t *testing.T) {
	namespacedTypeData = &NamespacedTypeData{}

	newP := namespacedTypeData.Add(resource1)
	versionMap, ok := newP.Mapping[testGroup1]
	assert.True(t, ok)
	assert.NotEmpty(t, versionMap)

	kindMap, ok := versionMap[testVersion1]
	assert.True(t, ok)
	assert.NotEmpty(t, kindMap)

	opMap, ok := kindMap[testKind1]
	assert.True(t, ok)
	assert.NotEmpty(t, opMap)

	instanceMap, ok := opMap[string(testOp1)]
	assert.True(t, ok)
	assert.NotEmpty(t, instanceMap)
	assert.True(t, instanceMap[uid1])
}

func TestDelete(t *testing.T) {
	namespacedTypeData = &NamespacedTypeData{}

	newP := namespacedTypeData.Add(resource1)
	newP = newP.Delete(resource1)

	versionMap, ok := newP.Mapping[testGroup1]
	assert.True(t, ok)
	assert.NotEmpty(t, versionMap)

	kindMap, ok := versionMap[testVersion1]
	assert.True(t, ok)
	assert.NotEmpty(t, kindMap)

	opMap, ok := kindMap[testKind1]
	assert.True(t, ok)
	assert.NotEmpty(t, opMap)

	instanceMap, ok := opMap[string(testOp1)]
	assert.True(t, ok)
	assert.Empty(t, instanceMap)
}

func TestUpdate(t *testing.T) {
	namespacedTypeData = &NamespacedTypeData{}
	newP := namespacedTypeData.Add(resource1)
	newP = newP.Update(resource1a)

	versionMap, ok := newP.Mapping[testGroup1]
	assert.True(t, ok)
	assert.NotEmpty(t, versionMap)

	kindMap, ok := versionMap[testVersion1]
	assert.True(t, ok)
	assert.NotEmpty(t, kindMap)

	opMap, ok := kindMap[testKind1]
	assert.True(t, ok)
	assert.NotEmpty(t, opMap)

	instanceMap, ok := opMap[string(testOp1)]
	assert.True(t, ok)
	assert.Empty(t, instanceMap)

	instanceMap, ok = opMap[string(testOp2)]
	assert.True(t, ok)
	assert.NotEmpty(t, instanceMap)
	assert.True(t, instanceMap[uid1])
}

func TestExist(t *testing.T) {
	namespacedTypeData = &NamespacedTypeData{}
	newP := namespacedTypeData.Add(resource1)

	gvk := &metav1.GroupVersionKind{
		Group:   testGroup1,
		Version: testVersion1,
		Kind:    testKind1,
	}

	assert.True(t, newP.Exist(gvk, testOp1))

	newP = newP.Add(resource2)

	assert.True(t, newP.Exist(gvk, testOp1))

	newP = newP.Delete(resource1)

	assert.True(t, newP.Exist(gvk, testOp1))

	newP = newP.Update(resource2a)
	assert.False(t, newP.Exist(gvk, testOp1))
	assert.True(t, newP.Exist(gvk, testOp2))
}

func TestGenerate(t *testing.T) {
	namespacedTypeData = &NamespacedTypeData{}
	namespacedTypeData = namespacedTypeData.Add(resource1)
	namespacedTypeData = namespacedTypeData.Add(resource2a)
	namespacedTypeData = namespacedTypeData.Add(resource3)

	config := namespacedTypeData.GenerateGlobalWebhook()
	assert.True(t, len(config.Webhooks) == 1)
	assert.True(t, len(config.Webhooks[0].Rules) == 2)

	if config.Webhooks[0].Rules[0].APIGroups[0] == testGroup2 {
		assert.Equal(t, config.Webhooks[0].Rules[0].APIGroups[0], testGroup2)
		assert.Equal(t, config.Webhooks[0].Rules[0].APIVersions[0], testVersion2)
		assert.Equal(t, config.Webhooks[0].Rules[0].Resources[0], testKind2)
		assert.Contains(t, config.Webhooks[0].Rules[0].Operations, testOp1)
		assert.Equal(t, config.Webhooks[0].Rules[1].APIGroups[0], testGroup1)
		assert.Equal(t, config.Webhooks[0].Rules[1].APIVersions[0], testVersion1)
		assert.Equal(t, config.Webhooks[0].Rules[1].Resources[0], testKind1)
		assert.Contains(t, config.Webhooks[0].Rules[1].Operations, testOp1)
		assert.Contains(t, config.Webhooks[0].Rules[1].Operations, testOp2)
	} else {
		assert.Equal(t, config.Webhooks[0].Rules[0].APIGroups[0], testGroup1)
		assert.Equal(t, config.Webhooks[0].Rules[0].APIVersions[0], testVersion1)
		assert.Equal(t, config.Webhooks[0].Rules[0].Resources[0], testKind1)
		assert.Contains(t, config.Webhooks[0].Rules[0].Operations, testOp1)
		assert.Contains(t, config.Webhooks[0].Rules[0].Operations, testOp2)
		assert.Equal(t, config.Webhooks[0].Rules[1].APIGroups[0], testGroup2)
		assert.Equal(t, config.Webhooks[0].Rules[1].APIVersions[0], testVersion2)
		assert.Equal(t, config.Webhooks[0].Rules[1].Resources[0], testKind2)
		assert.Contains(t, config.Webhooks[0].Rules[1].Operations, testOp1)
	}
}
