package namespacedvalidatingproxy

import (
	"bytes"
	"encoding/gob"

	appv1alpha1 "github.com/redislabs/gesher/pkg/apis/app/v1alpha1"
	"k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	EndpointData = &EndpointDataType{}
)

type WebhookConfig struct {
	ClientConfig  v1beta1.WebhookClientConfig
	FailurePolicy v1beta1.FailurePolicyType
	TimeoutSecs   int32
}

type typeInstanceMap map[types.UID]WebhookConfig
type typeOpMap map[v1beta1.OperationType]typeInstanceMap
type typeKindMap map[string]typeOpMap
type typeVersionMap map[string]typeKindMap
type typeGroupMap map[string]typeVersionMap
type typeNamespaceMap map[string]typeGroupMap

type EndpointDataType struct {
	Mapping typeNamespaceMap
}

func (p *EndpointDataType) Get(namespace string, kind *metav1.GroupVersionKind, op v1beta1.OperationType) []WebhookConfig {
	var ret []WebhookConfig

	if groupMap, ok := p.Mapping[namespace]; ok {
		groupList := []string{kind.Group, "*"}
		var versionMapList []typeVersionMap

		for _, group := range groupList {
			if versionMap, ok := groupMap[group]; ok {
				versionMapList = append(versionMapList, versionMap)
			}
		}

		versionList := []string{kind.Version, "*"}
		var kindMapList []typeKindMap
		for _, versionMap := range versionMapList {
			for _, version := range versionList {
				if kindMap, ok := versionMap[version]; ok {
					kindMapList = append(kindMapList, kindMap)
				}
			}
		}

		kindList := []string{kind.Kind, "*"}
		var opMapList []typeOpMap
		for _, kindMap := range kindMapList {
			for _, kind := range kindList {
				if opMap, ok := kindMap[kind]; ok {
					opMapList = append(opMapList, opMap)
				}
			}
		}

		opList := []v1beta1.OperationType{op, v1beta1.OperationAll}
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

func (p *EndpointDataType) Add(t *appv1alpha1.NamespacedValidatingProxy) *EndpointDataType {
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
			var kindMapList []typeKindMap
			for _, versionMap := range versionMapList {
				for _, version := range webhookRule.APIVersions {
					kindMap, ok := versionMap[version]
					if !ok {
						versionMap[version] = make(typeKindMap)
						kindMap = versionMap[version]
					}
					kindMapList = append(kindMapList, kindMap)
				}
			}
			var opMapList []typeOpMap
			for _, kindMap := range kindMapList {
				for _, kind := range webhookRule.Resources {
					opMap, ok := kindMap[kind]
					if !ok {
						kindMap[kind] = make(typeOpMap)
						opMap = kindMap[kind]
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

					var failurePolicy v1beta1.FailurePolicyType
					if webhook.FailurePolicy == nil {
						failurePolicy = v1beta1.Fail
					} else {
						failurePolicy = *webhook.FailurePolicy
					}
					var timeout int32
					if webhook.TimeoutSeconds == nil {
						timeout = 30
					} else {
						timeout = *webhook.TimeoutSeconds
					}
					instanceMap[t.UID] = WebhookConfig{
						ClientConfig:  webhook.ClientConfig,
						FailurePolicy: failurePolicy,
						TimeoutSecs:   timeout,
					}
				}
			}
		}
	}

	return newE
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

func (p *EndpointDataType) Delete(t *appv1alpha1.NamespacedValidatingProxy) *EndpointDataType {
	newE := copyEndpointData(p)

	if groupMap, ok := newE.Mapping[t.Namespace]; ok {
		for _, versionMap := range groupMap {
			for _, kindMap := range versionMap {
				for _, opMap := range kindMap {
					for _, instanceMap := range opMap {
						delete(instanceMap, t.UID)
					}
				}
			}
		}
	}

	return newE
}

func (p *EndpointDataType) Update(t *appv1alpha1.NamespacedValidatingProxy) *EndpointDataType {
	newE := p.Delete(t)
	newE = newE.Add(t)

	return newE
}

func (p *EndpointDataType) GenerateConfig() {
	// If we would generatean external config, perhaps stored in a config map
}
