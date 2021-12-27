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
	"bytes"
	"encoding/gob"

	admregv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	appv1alpha1 "github.com/redislabs/gesher/pkg/apis/app/v1alpha1"
)

var (
	EndpointData = &EndpointDataType{
		Mapping: make(typeNamespaceMap),
	}
)

type WebhookConfig struct {
	ClientConfig  admregv1.WebhookClientConfig
	FailurePolicy admregv1.FailurePolicyType
	TimeoutSecs   int32
}

type typeInstanceMap map[types.UID]WebhookConfig
type typeOpMap map[admregv1.OperationType]typeInstanceMap
type typeResourceMap map[string]typeOpMap
type typeVersionMap map[string]typeResourceMap
type typeGroupMap map[string]typeVersionMap
type typeNamespaceMap map[string]typeGroupMap

type EndpointDataType struct {
	Mapping typeNamespaceMap
}

func (p *EndpointDataType) Get(namespace string, resource metav1.GroupVersionResource, op admregv1.OperationType) []WebhookConfig {
	var ret []WebhookConfig

	if groupMap, ok := p.Mapping[namespace]; ok {
		groupList := []string{resource.Group, "*"}
		var versionMapList []typeVersionMap

		for _, group := range groupList {
			if versionMap, ok := groupMap[group]; ok {
				versionMapList = append(versionMapList, versionMap)
			}
		}

		versionList := []string{resource.Version, "*"}
		var resourceMapList []typeResourceMap
		for _, versionMap := range versionMapList {
			for _, version := range versionList {
				if resourceMap, ok := versionMap[version]; ok {
					resourceMapList = append(resourceMapList, resourceMap)
				}
			}
		}

		resourceList := []string{resource.Resource, "*"}
		var opMapList []typeOpMap
		for _, resourceMap := range resourceMapList {
			for _, resource := range resourceList {
				if opMap, ok := resourceMap[resource]; ok {
					opMapList = append(opMapList, opMap)
				}
			}
		}

		opList := []admregv1.OperationType{op, admregv1.OperationAll}
		var instanceMapList []typeInstanceMap
		for _, opMap := range opMapList {
			for _, op := range opList {
				if instanceMap, ok := opMap[op]; ok {
					instanceMapList = append(instanceMapList, instanceMap)
				}
			}
		}

		for _, instanceMap := range instanceMapList {
			for _, webhookConfig := range instanceMap {
				ret = append(ret, webhookConfig)
			}
		}
	}

	return ret
}

func (p *EndpointDataType) Add(t *appv1alpha1.NamespacedValidatingRule) *EndpointDataType {
	newE := copyEndpointData(p)

	if newE.Mapping == nil {
		newE.Mapping = make(typeNamespaceMap)
	}

	namespaceMap := newE.Mapping
	if _, ok := namespaceMap[t.Namespace]; !ok {
		namespaceMap[t.Namespace] = make(typeGroupMap)
	}

	groupMap := namespaceMap[t.Namespace]

	for _, webhook := range t.Spec.Webhooks {
		webhookConfig := createWebhookConfig(webhook, t.Namespace)

		for _, webhookRule := range webhook.Rules {
			var versionMapList []typeVersionMap
			for _, group := range webhookRule.APIGroups {
				versionMap, ok := groupMap[group]
				if !ok {
					groupMap[group] = make(typeVersionMap)
					versionMap = groupMap[group]
				}
				versionMapList = append(versionMapList, versionMap)
			}
			var resourceMapList []typeResourceMap
			for _, versionMap := range versionMapList {
				for _, version := range webhookRule.APIVersions {
					resourceMap, ok := versionMap[version]
					if !ok {
						versionMap[version] = make(typeResourceMap)
						resourceMap = versionMap[version]
					}
					resourceMapList = append(resourceMapList, resourceMap)
				}
			}
			var opMapList []typeOpMap
			for _, resourceMap := range resourceMapList {
				for _, resource := range webhookRule.Resources {
					opMap, ok := resourceMap[resource]
					if !ok {
						resourceMap[resource] = make(typeOpMap)
						opMap = resourceMap[resource]
					}
					opMapList = append(opMapList, opMap)
				}
			}

			for _, opMap := range opMapList {
				for _, op := range webhookRule.Operations {
					instanceMap, ok := opMap[op]
					if !ok {
						opMap[op] = make(typeInstanceMap)
						instanceMap = opMap[op]
					}

					instanceMap[t.UID] = webhookConfig
				}
			}
		}
	}

	return newE
}

func createWebhookConfig(webhook admregv1.ValidatingWebhook, namespace string) WebhookConfig {
	var (
		failurePolicy admregv1.FailurePolicyType
		timeout       int32
	)

	if webhook.FailurePolicy == nil {
		failurePolicy = admregv1.Fail
	} else {
		failurePolicy = *webhook.FailurePolicy
	}

	if webhook.TimeoutSeconds == nil {
		timeout = 30
	} else {
		timeout = *webhook.TimeoutSeconds
	}

	if webhook.ClientConfig.Service != nil && webhook.ClientConfig.Service.Namespace == "" {
		webhook.ClientConfig.Service.Namespace = namespace
	}

	return WebhookConfig{
		ClientConfig:  webhook.ClientConfig,
		FailurePolicy: failurePolicy,
		TimeoutSecs:   timeout,
	}
}

func copyEndpointData(p *EndpointDataType) *EndpointDataType {
	var newP EndpointDataType

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)

	err := enc.Encode(p)
	if err != nil {
		return nil
	}

	err = dec.Decode(&newP)
	if err != nil {
		return nil
	}

	return &newP
}

func (p *EndpointDataType) Delete(t *appv1alpha1.NamespacedValidatingRule) *EndpointDataType {
	newE := copyEndpointData(p)

	if groupMap, ok := newE.Mapping[t.Namespace]; ok {
		for _, versionMap := range groupMap {
			for _, resourceMap := range versionMap {
				for _, opMap := range resourceMap {
					for _, instanceMap := range opMap {
						delete(instanceMap, t.UID)
					}
				}
			}
		}
	}

	return newE
}

func (p *EndpointDataType) Update(t *appv1alpha1.NamespacedValidatingRule) *EndpointDataType {
	newE := p.Delete(t)
	newE = newE.Add(t)

	return newE
}

func (p *EndpointDataType) GenerateConfig() {
	// If we would generate an external config for use by an external proxy pod, perhaps stored in a config map
}
