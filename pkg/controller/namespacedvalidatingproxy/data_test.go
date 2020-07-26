package namespacedvalidatingproxy

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

	testGroup1   = "testGroup1"
	testVersion1 = "testVersion1"
	testKind1    = "testKind1"
	testOp1      = v1beta1.Create
	testOp2      = v1beta1.Delete
)

var (
	resource1 = &v1alpha1.NamespacedValidatingProxy{
		ObjectMeta: metav1.ObjectMeta{
			UID:       uid1,
			Namespace: namespace,
		},
		Spec: v1alpha1.NamespacedValidatingProxySpec{
			Webhooks: []v1beta1.ValidatingWebhook{{
				Name:         "resource1",
				ClientConfig: v1beta1.WebhookClientConfig{},
				Rules: []v1beta1.RuleWithOperations{{
					Operations: []v1beta1.OperationType{testOp1},
					Rule: v1beta1.Rule{
						APIGroups:   []string{testGroup1},
						APIVersions: []string{testVersion1},
						Resources:   []string{testKind1},
					},
				}},
			}},
		},
	}

	resource1a = &v1alpha1.NamespacedValidatingProxy{
		ObjectMeta: metav1.ObjectMeta{
			UID:       uid1,
			Namespace: namespace,
		},
		Spec: v1alpha1.NamespacedValidatingProxySpec{
			Webhooks: []v1beta1.ValidatingWebhook{{
				Name:         "resource1",
				ClientConfig: v1beta1.WebhookClientConfig{},
				Rules: []v1beta1.RuleWithOperations{{
					Operations: []v1beta1.OperationType{testOp2},
					Rule: v1beta1.Rule{
						APIGroups:   []string{testGroup1},
						APIVersions: []string{testVersion1},
						Resources:   []string{testKind1},
					},
				}},
			}},
		},
	}

	resource2 = &v1alpha1.NamespacedValidatingProxy{
		ObjectMeta: metav1.ObjectMeta{
			UID:       uid2,
			Namespace: namespace,
		},
		Spec: v1alpha1.NamespacedValidatingProxySpec{
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
						Resources:   []string{testKind1},
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

	kindMap, ok := versionMap[testVersion1]
	assert.True(t, ok)
	assert.NotEmpty(t, kindMap)

	opMap, ok := kindMap[testKind1]
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

	kindMap, ok := versionMap[testVersion1]
	assert.True(t, ok)
	assert.NotEmpty(t, kindMap)

	opMap, ok := kindMap[testKind1]
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

	kindMap, ok := versionMap[testVersion1]
	assert.True(t, ok)
	assert.NotEmpty(t, kindMap)

	opMap, ok := kindMap[testKind1]
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
	w := newE.Get(namespace, &metav1.GroupVersionKind{Group: testGroup1, Version: testVersion1, Kind: testKind1}, testOp1)
	assert.NotEmpty(t, w)
	assert.Len(t, w, 1)
	assert.Equal(t, w[0].ClientConfig.Service.Namespace, namespace)
}
