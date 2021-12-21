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

package admission_tester_test

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/redislabs/gesher/integration-tests/common"
	admregv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAdmissionTester(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AdmissionTester Suite")
}

var (
	sa                 *corev1.ServiceAccount
	role               *rbacv1.Role
	roleBinding        *rbacv1.RoleBinding
	clusterRole        *rbacv1.ClusterRole
	clusterRoleBinding *rbacv1.ClusterRoleBinding

	webhook *admregv1.ValidatingWebhookConfiguration
	service *corev1.Service
	secret  *corev1.Secret
	deploy  *appsv1.Deployment

	kubeClient client.Client
)

var _ = BeforeSuite(func() {
	var err error
	kubeClient, _, err = common.GetClient()
	Expect(err).To(Succeed())

	sa = common.LoadServiceAccount()
	role = common.LoadRole()
	roleBinding = common.LoadRoleBinding()
	clusterRole = common.LoadClusterRole()
	clusterRoleBinding = common.LoadClusterRoleBinding()

	service = common.LoadTestService()
	deploy = common.LoadAdmissionDeploy()

	var s corev1.Secret
	Expect(kubeClient.Get(context.TODO(), types.NamespacedName{Name: "admission-test", Namespace: common.Namespace}, &s)).To(Succeed())
	secret = &s

	webhook = loadWebhook(secret)
})

var _ = AfterSuite(func() {
	if webhook != nil {
		Expect(kubeClient.Delete(context.TODO(), webhook)).To(Succeed())
		webhook = nil
	}
	if service != nil {
		Expect(kubeClient.Delete(context.TODO(), service)).To(Succeed())
		service = nil
	}
	if deploy != nil {
		Expect(kubeClient.Delete(context.TODO(), deploy)).To(Succeed())
		deploy = nil
	}
	if secret != nil {
		Expect(kubeClient.Delete(context.TODO(), secret)).To(Succeed())
		secret = nil
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

func loadWebhook(s *corev1.Secret) *admregv1.ValidatingWebhookConfiguration {
	By("Read and Load Webhook")

	path := "/admission"
	failurePolicy := admregv1.Fail
	scope := admregv1.AllScopes
	sideEffects := admregv1.SideEffectClassNone

	webhook := &admregv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "admission-test",
		},
		Webhooks: []admregv1.ValidatingWebhook{
			{
				Name: "test.admission.gesher",
				ClientConfig: admregv1.WebhookClientConfig{
					Service: &admregv1.ServiceReference{
						Name:      "admission-test",
						Namespace: common.Namespace,
						Path:      &path,
					},
					CABundle: s.Data["cert"],
				},
				Rules: []admregv1.RuleWithOperations{
					{
						Operations: []admregv1.OperationType{admregv1.Create},
						Rule: admregv1.Rule{
							APIGroups:   []string{""},
							APIVersions: []string{"v1"},
							Resources:   []string{"namespaces"},
							Scope:       &scope,
						},
					},
				},
				FailurePolicy:           &failurePolicy,
				SideEffects:             &sideEffects,
				AdmissionReviewVersions: []string{"v1"},
			},
		},
	}

	Expect(kubeClient.Create(context.TODO(), webhook)).To(Succeed())

	return webhook
}
