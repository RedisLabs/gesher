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
	"fmt"
	"time"

	admissionTest "github.com/redislabs/gesher/pkg/admission-test"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/redislabs/gesher/integration-tests/common"
	"github.com/redislabs/gesher/pkg/apis/app/v1alpha1"
)

var _ = Describe("NamespacedWebhook", func() {
	var (
		pod            *corev1.Pod
		namespacedType *v1alpha1.NamespacedValidatingType
		webhook        *v1alpha1.NamespacedValidatingRule
	)

	AfterEach(func() {
		if CurrentGinkgoTestDescription().Failed {
			fmt.Fprintf(GinkgoWriter, "failure: FIXME: log collection would go here, before any teardown!")
		}

		if pod != nil {
			Expect(kubeClient.Delete(context.TODO(), pod)).To(Succeed())
			pod = nil
		}

		if namespacedType != nil {
			Expect(kubeClient.Delete(context.TODO(), namespacedType)).To(Succeed())
			Eventually(func() error { return common.VerifyDeleted(namespacedType) }, 60, 5).Should(Succeed())
			namespacedType = nil
		}

		if webhook != nil {
			Expect(kubeClient.Delete(context.TODO(), webhook)).To(Succeed())
			Eventually(func() error { return common.VerifyDeleted(webhook) }, 60, 5).Should(Succeed())
			webhook = nil
		}

		// This has to be at the end after webhook's are deleted, as otherwise webhook can block pod from loading
		if admDeploy == nil {
			admDeploy = common.LoadAdmissionDeploy()
		}
	})

	It("Create Pod without NamespacedType", func() {
		pod = getPod(false)
		tryCreatePod(&pod, true, true)
	})

	It("Create Pod with NamespacedType, but without Namespaced Rule", func() {
		namespacedType = createNamespacedType()

		pod = getPod(false)
		tryCreatePod(&pod, true, true)
	})

	It("Create Pod with Namespaced Webhook, without label", func() {
		namespacedType = createNamespacedType()
		webhook = createNamespacedWebhook(admissionv1.Fail)

		pod = getPod(false)
		tryCreatePod(&pod, false, true)
	})

	It("Create Pod with Namespaced Webhook, with label", func() {
		namespacedType = createNamespacedType()
		webhook = createNamespacedWebhook(admissionv1.Fail)

		pod = getPod(true)
		tryCreatePod(&pod, true, true)
	})

	It("Create Pod after delete Namespaced Webhook, without label", func() {
		namespacedType = createNamespacedType()
		webhook = createNamespacedWebhook(admissionv1.Fail)
		webhook = deleteNamespacedWebhook(webhook)

		pod = getPod(false)
		tryCreatePod(&pod, true, true)
	})

	It("Test with Ignore Failure type if admission controller is removed", func() {
		namespacedType = createNamespacedType()
		webhook = createNamespacedWebhook(admissionv1.Ignore)

		Expect(kubeClient.Delete(context.TODO(), admDeploy)).To(Succeed())
		admDeploy = nil
		Eventually(func() error { return common.VerifyNoEndpoint(admService.Name, admService.Namespace) }, 60, 5).Should(Succeed())

		pod = getPod(false)
		tryCreatePod(&pod, true, false)
	})

	It("Test with Fail Failure type if admission controller is removed", func() {
		namespacedType = createNamespacedType()
		webhook = createNamespacedWebhook(admissionv1.Fail)

		Expect(kubeClient.Delete(context.TODO(), admDeploy)).To(Succeed())
		admDeploy = nil
		Eventually(func() error { return common.VerifyNoEndpoint(admService.Name, admService.Namespace) }, 60, 5).Should(Succeed())

		pod = getPod(false)
		tryCreatePod(&pod, false, false)
	})
})

func createNamespacedType() *v1alpha1.NamespacedValidatingType {
	By("Add NamespacedValidatingType")
	pt := &v1alpha1.NamespacedValidatingType{
		ObjectMeta: metav1.ObjectMeta{
			Name: "gesher-test",
		},
		Spec: v1alpha1.NamespacedValidatingTypeSpec{
			Types: []admissionv1.RuleWithOperations{{
				Operations: []admissionv1.OperationType{"CREATE"},
				Rule: admissionv1.Rule{
					APIGroups:   []string{""},
					APIVersions: []string{"v1"},
					Resources:   []string{"pods"},
				},
			}},
		},
	}
	Expect(kubeClient.Create(context.TODO(), pt)).To(Succeed())

	By("wait on resource to be applied")
	Eventually(func() error { return common.VerifyApplied(pt) }, 60, 5).Should(Succeed())

	return pt
}

func getPod(label bool) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("pod-test-%v", time.Now().Unix()),
			Namespace: common.Namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "pod-test",
					Image: "kubernetes/pause:latest",
				},
			},
		},
	}

	if label {
		pod.Labels = map[string]string{admissionTest.AdmissionKey: admissionTest.AdmissionAllow}
	}

	return pod
}

// Function does a little pointer setting to set *pod pointer to be nil if object wasn't created.
// needs to be complicated, as also has to handle creations that shouldn't work, but do (i.e. a test failure)
func tryCreatePod(pod **corev1.Pod, success bool, available bool) {
	switch success {
	case true:
		tmp := *pod
		*pod = nil
		Expect(kubeClient.Create(context.TODO(), tmp)).To(Succeed())
		*pod = tmp
	case false:
		switch available {
		case true:
			err := kubeClient.Create(context.TODO(), *pod)
			if err != nil {
				*pod = nil
			}
			Expect(err).To(MatchError(MatchRegexp("admission-allow label not set to allow this resource")))
		case false:
			err := kubeClient.Create(context.TODO(), *pod)
			if err != nil {
				*pod = nil
			}
			Expect(err).To(MatchError(MatchRegexp("connect: connection refused")))
		}
		*pod = nil
	}
}

func createNamespacedWebhook(failurePolicy admissionv1.FailurePolicyType) *v1alpha1.NamespacedValidatingRule {
	By("Add NamespacedValidatingRule")
	path := "/admission"
	sideEffect := admissionv1.SideEffectClassNone

	nvp := &v1alpha1.NamespacedValidatingRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-hook",
			Namespace: common.Namespace,
		},
		Spec: v1alpha1.NamespacedValidatingRuleSpec{
			Webhooks: []admissionv1.ValidatingWebhook{
				{
					Name:          "test-hook",
					FailurePolicy: &failurePolicy,
					ClientConfig: admissionv1.WebhookClientConfig{
						Service: &admissionv1.ServiceReference{
							Name:      admService.Name,
							Namespace: admService.Namespace,
							Port:      &admService.Spec.Ports[0].Port,
							Path:      &path,
						},
						CABundle: admSecret.Data["cert"],
					},
					Rules: []admissionv1.RuleWithOperations{
						{
							Operations: []admissionv1.OperationType{admissionv1.Create},
							Rule: admissionv1.Rule{
								APIGroups:   []string{""},
								APIVersions: []string{"v1"},
								Resources:   []string{"pods"},
							},
						},
					},
					SideEffects:             &sideEffect,
					AdmissionReviewVersions: []string{"v1"},
				},
			},
		},
	}

	Expect(kubeClient.Create(context.TODO(), nvp)).To(Succeed())

	By("wait on resource to be applied")
	Eventually(func() error { return common.VerifyApplied(nvp) }, 60, 5).Should(Succeed())

	return nvp
}

func deleteNamespacedWebhook(nvp *v1alpha1.NamespacedValidatingRule) *v1alpha1.NamespacedValidatingRule {
	Expect(kubeClient.Delete(context.TODO(), nvp)).To(Succeed())
	Eventually(func() error { return common.VerifyDeleted(nvp) }, 60, 5).Should(Succeed())

	return nil
}
