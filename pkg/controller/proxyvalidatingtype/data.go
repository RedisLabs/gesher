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

package proxyvalidatingtype

import (
	"bytes"
	"encoding/gob"
	"github.com/redislabs/gesher/cmd/manager/flags"

	"k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	appv1alpha1 "github.com/redislabs/gesher/pkg/apis/app/v1alpha1"
)

var (
	proxyTypeData = &ProxyTypeData{}
	caBundle      []byte
)

type typeInstanceMap map[types.UID]bool
type typeOpMap map[string]typeInstanceMap
type typeKindMap map[string]typeOpMap
type typeVersionMap map[string]typeKindMap
type typeGroupMap map[string]typeVersionMap

type ProxyTypeData struct {
	Mapping typeGroupMap
}

func (p *ProxyTypeData) Exist(kind *metav1.GroupVersionKind, op v1beta1.OperationType) bool {
	groupList := []string{kind.Group, "*"}
	var versionMapList []typeVersionMap
	for _, group := range groupList {
		if versionMap, ok := p.Mapping[group]; ok {
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

	opList := []string{string(op), "*"}
	for _, opMap := range opMapList {
		for _, op := range opList {
			if opList, ok := opMap[op]; ok {
				if len(opList) > 0 {
					return true
				}
			}
		}
	}

	return false
}

func (p *ProxyTypeData) Add(t *appv1alpha1.ProxyValidatingType) *ProxyTypeData {
	newP := copyProxyTypeData(p)

	if newP.Mapping == nil {
		newP.Mapping = make(typeGroupMap)
	}

	groupMap := newP.Mapping

	for _, proxyType := range t.Spec.Types {
		var versionMapList []typeVersionMap
		for _, group := range proxyType.APIGroups {
			versionMap, ok := groupMap[group]
			if !ok {
				groupMap[group] = make(typeVersionMap)
				versionMap = groupMap[group]
			}
			versionMapList = append(versionMapList, versionMap)
		}
		var kindMapList []typeKindMap
		for _, versionMap := range versionMapList {
			for _, version := range proxyType.APIVersions {
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
			for _, kind := range proxyType.Resources {
				opMap, ok := kindMap[kind]
				if !ok {
					kindMap[kind] = make(typeOpMap)
					opMap = kindMap[kind]
				}
				opMapList = append(opMapList, opMap)
			}
		}

		for _, opMap := range opMapList {
			for _, op := range proxyType.Operations {
				opMap[string(op)] = map[types.UID]bool{t.UID: true}
			}
		}
	}

	return newP
}

func copyProxyTypeData(p *ProxyTypeData) *ProxyTypeData {
	var newP ProxyTypeData

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

func (p *ProxyTypeData) Delete(t *appv1alpha1.ProxyValidatingType) *ProxyTypeData {
	newP := copyProxyTypeData(p)

	for _, versionMap := range newP.Mapping {
		for _, kindMap := range versionMap {
			for _, opMap := range kindMap {
				for _, instanceMap := range opMap {
					delete(instanceMap, t.UID)
				}
			}
		}
	}

	return newP
}

func (p *ProxyTypeData) Update(t *appv1alpha1.ProxyValidatingType) *ProxyTypeData {
	newP := p.Delete(t)
	newP = newP.Add(t)

	return newP
}

func (p *ProxyTypeData) GenerateGlobalWebhook() *v1beta1.ValidatingWebhookConfiguration {
	webhook := &v1beta1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: ProxyWebhookName},
	}

	webhook.Webhooks = p.enumerateWebhooks()

	return webhook
}

func (p *ProxyTypeData) enumerateWebhooks() []v1beta1.ValidatingWebhook {
	var rules []v1beta1.RuleWithOperations

	scope := v1beta1.NamespacedScope

	for group, versionMap := range p.Mapping {
		for version, kindMap := range versionMap {
			for kind, opMap := range kindMap {
				var opList []v1beta1.OperationType
				for op, instanceMap := range opMap {
					if len(instanceMap) > 0 {
						switch op {
						case string(v1beta1.OperationAll):
							opList = append(opList, v1beta1.OperationAll)
						case string(v1beta1.Create):
							opList = append(opList, v1beta1.Create)
						case string(v1beta1.Update):
							opList = append(opList, v1beta1.Update)
						case string(v1beta1.Delete):
							opList = append(opList, v1beta1.Delete)
						case string(v1beta1.Connect):
							opList = append(opList, v1beta1.Connect)
						}
					}
				}
				if len(opList) > 0 {
					rule := v1beta1.RuleWithOperations{
						Rule: v1beta1.Rule{
							APIGroups:   []string{group},
							APIVersions: []string{version},
							Resources:   []string{kind},
							Scope:       &scope,
						},
						Operations: opList,
					}
					rules = append(rules, rule)
				}
			}
		}
	}

	fail := v1beta1.Fail
	var defaultTimeout int32 = 30
	sideEffects := v1beta1.SideEffectClassUnknown
	webhook := v1beta1.ValidatingWebhook{
		Name:                    ProxyWebhookName,
		ClientConfig:            selfConfig(),
		Rules:                   rules,
		FailurePolicy:           &fail,
		SideEffects:             &sideEffects,
		NamespaceSelector:       &metav1.LabelSelector{},
		TimeoutSeconds:          &defaultTimeout,
		AdmissionReviewVersions: []string{"v1beta1"},
	}

	return []v1beta1.ValidatingWebhook{webhook}
}

// FiXME
func selfConfig() v1beta1.WebhookClientConfig {
	path := "/proxy"

	return v1beta1.WebhookClientConfig{
		Service: &v1beta1.ServiceReference{
			Namespace: *flags.Namespace,
			Name:      *flags.Service,
			Path:      &path,
		},
		CABundle: caBundle,
	}
}
