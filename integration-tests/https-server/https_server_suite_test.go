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

package https_server_test

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/redislabs/gesher/integration-tests/common"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestHttpsServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HttpsServer Suite")
}

var (
	kubeClient  client.Client
	serviceName string

	crd1               *apiextv1.CustomResourceDefinition
	crd2               *apiextv1.CustomResourceDefinition
	service            *corev1.Service
	sa                 *corev1.ServiceAccount
	role               *rbacv1.Role
	roleBinding        *rbacv1.RoleBinding
	clusterRole        *rbacv1.ClusterRole
	clusterRoleBinding *rbacv1.ClusterRoleBinding
)

var _ = BeforeSuite(func() {
	var err error

	By("Get client")
	kubeClient, _, err = common.GetClient()
	Expect(err).To(Succeed())

	crd1 = common.LoadNamespacedValidatingTypeCRD()
	crd2 = common.LoadNamespacedValidatingRuleCRD()
	service = common.LoadService()
	serviceName = service.Name
	sa = common.LoadServiceAccount()
	clusterRole = common.LoadClusterRole()
	clusterRoleBinding = common.LoadClusterRoleBinding()
	role = common.LoadRole()
	roleBinding = common.LoadRoleBinding()
})

var _ = AfterSuite(func() {
	if crd1 != nil {
		Expect(kubeClient.Delete(context.TODO(), crd1)).To(Succeed())
		crd1 = nil
	}
	if crd2 != nil {
		Expect(kubeClient.Delete(context.TODO(), crd2)).To(Succeed())
		crd1 = nil
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
})
