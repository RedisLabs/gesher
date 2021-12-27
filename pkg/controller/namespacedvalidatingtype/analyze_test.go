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
	"testing"

	"github.com/stretchr/testify/assert"
	admregv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	appv1alpha1 "github.com/redislabs/gesher/pkg/apis/app/v1alpha1"
)

var (
	logger = zap.New()
)

const (
	uid         = "1"
	testOp      = admregv1.Create
	testDiffOp  = admregv1.Delete
	testGroup   = "testGroup"
	testVersion = "testVersion"
	testKind    = "testKind"
)

var (
	rule = admregv1.Rule{
		APIGroups:   []string{testGroup},
		APIVersions: []string{testVersion},
		Resources:   []string{testKind},
	}
)

func TestAnalyzeSame(t *testing.T) {
	namespacedTypeData := &NamespacedTypeData{}
	customResource := &appv1alpha1.NamespacedValidatingType{
		ObjectMeta: metav1.ObjectMeta{UID: uid},
		Spec: appv1alpha1.NamespacedValidatingTypeSpec{
			Types: []admregv1.RuleWithOperations{{
				Operations: []admregv1.OperationType{testOp},
				Rule:       rule,
			}},
		},
	}

	namespacedTypeData = namespacedTypeData.Add(customResource)
	webhook := namespacedTypeData.GenerateGlobalWebhook()

	observed := &observedState{
		customResource: customResource,
		clusterWebhook: webhook,
	}

	state, err := analyze(observed, logger)
	assert.Nil(t, err)
	assert.False(t, state.update)
}

func TestAnalyzeDifferent(t *testing.T) {
	namespacedTypeData := &NamespacedTypeData{}
	customResource := &appv1alpha1.NamespacedValidatingType{
		ObjectMeta: metav1.ObjectMeta{UID: uid},
		Spec: appv1alpha1.NamespacedValidatingTypeSpec{
			Types: []admregv1.RuleWithOperations{{
				Operations: []admregv1.OperationType{testOp},
				Rule:       rule,
			}},
		},
	}

	namespacedTypeData = namespacedTypeData.Add(customResource)
	webhook := namespacedTypeData.GenerateGlobalWebhook()
	customResource.Spec.Types[0].Operations[0] = testDiffOp

	observed := &observedState{
		customResource: customResource,
		clusterWebhook: webhook,
	}

	state, err := analyze(observed, logger)
	assert.Nil(t, err)
	assert.True(t, state.update)
}
