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

package namespaced_webhook_test

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redislabs/gesher/cmd/manager/flags"
	"github.com/redislabs/gesher/integration-tests/common"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func TestNamespacedWebhook(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NamespacedWebhook Suite")
}

var (
	crd1        *v1beta1.CustomResourceDefinition
	crd2        *v1beta1.CustomResourceDefinition
	opDeploy    *appsv1.Deployment
	admDeploy   *appsv1.Deployment
	sa          *corev1.ServiceAccount
	opService   *corev1.Service
	admService  *corev1.Service
	role        *rbacv1beta1.ClusterRole
	roleBinding *rbacv1beta1.ClusterRoleBinding
	opSecret    *corev1.Secret
	admSecret   *corev1.Secret

	kubeClient client.Client
)

var _ = BeforeSuite(func() {
	var err error

	By("Setup kube clients")
	kubeClient, _, err = common.GetClient()
	Expect(err).To(Succeed())

	crd1 = common.LoadNamespacedValidatingTypeCRD()
	crd2 = common.LoadNamespacedValidatingRuleCRD()
	opService = common.LoadService()
	admService = common.LoadTestService()

	sa = common.LoadServiceAccount()
	role = common.LoadClusterRole()
	roleBinding = common.LoadClusterRoleBinding()
	opDeploy = common.LoadOperator("Read and Load Operator")
	admDeploy = common.LoadAdmissionDeploy()

	var s corev1.Secret
	Expect(kubeClient.Get(context.TODO(), types.NamespacedName{Name: flags.DefaultTlsSecret, Namespace: common.Namespace}, &s)).To(Succeed())
	opSecret = &s

	var s1 corev1.Secret
	Expect(kubeClient.Get(context.TODO(), types.NamespacedName{Name: "admission-test", Namespace: common.Namespace}, &s1)).To(Succeed())
	admSecret = &s1
})

var _ = AfterSuite(func() {
	if opDeploy != nil {
		Expect(kubeClient.Delete(context.TODO(), opDeploy)).To(Succeed())
		opDeploy = nil
	}

	if crd1 != nil {
		Expect(kubeClient.Delete(context.TODO(), crd1)).To(Succeed())
		crd1 = nil
	}

	if crd2 != nil {
		Expect(kubeClient.Delete(context.TODO(), crd2)).To(Succeed())
		crd2 = nil
	}

	if opService != nil {
		Expect(kubeClient.Delete(context.TODO(), opService)).To(Succeed())
		opService = nil
	}

	if admService != nil {
		Expect(kubeClient.Delete(context.TODO(), admService)).To(Succeed())
		admService = nil
	}


	if sa != nil {
		Expect(kubeClient.Delete(context.TODO(), sa)).To(Succeed())
		sa = nil
	}

	if role != nil {
		Expect(kubeClient.Delete(context.TODO(), role)).To(Succeed())
		role = nil
	}

	if roleBinding != nil {
		Expect(kubeClient.Delete(context.TODO(), roleBinding)).To(Succeed())
		roleBinding = nil
	}

	if opSecret != nil {
		Expect(kubeClient.Delete(context.TODO(), opSecret)).To(Succeed())
		opSecret = nil
	}

	if admDeploy != nil {
		Expect(kubeClient.Delete(context.TODO(), admDeploy)).To(Succeed())
		admDeploy = nil
	}

	if admSecret != nil {
		Expect(kubeClient.Delete(context.TODO(), admSecret)).To(Succeed())
		admSecret = nil
	}
})

