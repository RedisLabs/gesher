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

package type_controller_test

import (
	"context"
	"testing"

	"github.com/redislabs/gesher/cmd/manager/flags"
	"github.com/redislabs/gesher/integration-tests/common"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"

	appsv1 "k8s.io/api/apps/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTypeController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TypeController Suite")
}

var (
	crd1               *apiextv1.CustomResourceDefinition
	crd2               *apiextv1.CustomResourceDefinition
	deploy             *appsv1.Deployment
	sa                 *corev1.ServiceAccount
	service            *corev1.Service
	role               *rbacv1.Role
	roleBinding        *rbacv1.RoleBinding
	clusterRole        *rbacv1.ClusterRole
	clusterRoleBinding *rbacv1.ClusterRoleBinding
	secret             *corev1.Secret

	kubeClient client.Client
)

var _ = BeforeSuite(func() {
	var err error

	By("Setup kube clients")
	kubeClient, _, err = common.GetClient()
	Expect(err).To(Succeed())

	crd1 = common.LoadNamespacedValidatingTypeCRD()
	crd2 = common.LoadNamespacedValidatingRuleCRD()
	service = common.LoadService()
	sa = common.LoadServiceAccount()
	role = common.LoadRole()
	roleBinding = common.LoadRoleBinding()
	clusterRole = common.LoadClusterRole()
	clusterRoleBinding = common.LoadClusterRoleBinding()
	deploy = common.LoadOperator("Read and Load Operator")

	var s corev1.Secret
	Expect(kubeClient.Get(context.TODO(), types.NamespacedName{Name: flags.DefaultTlsSecret, Namespace: common.Namespace}, &s)).To(Succeed())
	secret = &s
})

var _ = AfterSuite(func() {
	// unload pod running operator
	if deploy != nil {
		Expect(kubeClient.Delete(context.TODO(), deploy)).To(Succeed())
		deploy = nil
	}

	// unload CRD
	if crd1 != nil {
		Expect(kubeClient.Delete(context.TODO(), crd1)).To(Succeed())
		crd1 = nil
	}

	if crd2 != nil {
		Expect(kubeClient.Delete(context.TODO(), crd2)).To(Succeed())
		crd2 = nil
	}

	if service != nil {
		Expect(kubeClient.Delete(context.TODO(), service)).To(Succeed())
		service = nil
	}

	if sa != nil {
		Expect(kubeClient.Delete(context.TODO(), sa)).To(Succeed())
		sa = nil
	}

	if clusterRole != nil {
		Expect(kubeClient.Delete(context.TODO(), clusterRole)).To(Succeed())
		clusterRole = nil
	}

	if clusterRoleBinding != nil {
		Expect(kubeClient.Delete(context.TODO(), clusterRoleBinding)).To(Succeed())
		clusterRoleBinding = nil
	}

	if role != nil {
		Expect(kubeClient.Delete(context.TODO(), role)).To(Succeed())
		role = nil
	}

	if roleBinding != nil {
		Expect(kubeClient.Delete(context.TODO(), roleBinding)).To(Succeed())
		roleBinding = nil
	}

	if secret != nil {
		Expect(kubeClient.Delete(context.TODO(), secret)).To(Succeed())
		secret = nil
	}
})
