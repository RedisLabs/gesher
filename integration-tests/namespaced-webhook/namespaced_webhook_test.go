package namespaced_webhook_test

import (
	"context"
	"fmt"
	admissionTest "github.com/redislabs/gesher/pkg/admission-test"
	admissionv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/redislabs/gesher/integration-tests/common"
	"github.com/redislabs/gesher/pkg/apis/app/v1alpha1"
)

var _ = Describe("NamespacedWebhook", func() {
	var (
		pod       *corev1.Pod
		proxyType *v1alpha1.ProxyValidatingType
		webhook   *v1alpha1.NamespacedValidatingProxy
	)

	AfterEach(func() {
		if CurrentGinkgoTestDescription().Failed {
        	fmt.Fprintf(GinkgoWriter, "failure: FIXME: log collection would go here, before any teardown!")
    	}

		if pod != nil {
			Expect(kubeClient.Delete(context.TODO(), pod)).To(Succeed())
			pod = nil
		}

		if proxyType != nil {
			Expect(kubeClient.Delete(context.TODO(), proxyType)).To(Succeed())
			Eventually(func () error { return common.VerifyDeleted(proxyType)}, 60, 5).Should(Succeed())
			proxyType = nil
		}

		if webhook != nil {
			Expect(kubeClient.Delete(context.TODO(), webhook)).To(Succeed())
			Eventually(func () error { return common.VerifyDeleted(webhook)}, 60, 5).Should(Succeed())
			webhook = nil
		}

		// This has to be at the end after webhook's are deleted, as otherwise webhook can block pod from loading
		if admDeploy == nil {
			admDeploy = common.LoadAdmissionDeploy()
		}
	})

	It("Create Pod without ProxyType", func() {
		pod = getPod(false)
		tryCreatePod(&pod, true, true)
	})

	It("Create Pod with ProxyType, but without Namespaced Webhook", func() {
		proxyType = createProxyType()

		pod = getPod(false)
		tryCreatePod(&pod, true, true)
	})

	It("Create Pod with Namespaced Webhook, without label", func() {
		proxyType = createProxyType()
		webhook = createNamespacedWebhook(admissionv1beta1.Fail)

		pod = getPod(false)
		tryCreatePod(&pod, false, true)
	})

	It("Create Pod with Namespaced Webhook, with label", func() {
		proxyType = createProxyType()
		webhook = createNamespacedWebhook(admissionv1beta1.Fail)

		pod = getPod(true)
		tryCreatePod(&pod, true, true)
	})

	It("Create Pod after delete Namespaced Webhook, without label", func() {
		proxyType = createProxyType()
		webhook = createNamespacedWebhook(admissionv1beta1.Fail)
		webhook = deleteNamespacedWebhook(webhook)

		pod = getPod(false)
		tryCreatePod(&pod, true, true)
	})

	It("Test with Ignore Failure type if admission controller is removed", func() {
		proxyType = createProxyType()
		webhook = createNamespacedWebhook(admissionv1beta1.Ignore)

		Expect(kubeClient.Delete(context.TODO(), admDeploy)).To(Succeed())
		admDeploy = nil
		Eventually(func () error { return common.VerifyNoEndpoint(admService.Name, admService.Namespace) }, 60, 5).Should(Succeed())

		pod = getPod(false)
		tryCreatePod(&pod, true, false)
	})

	It("Test with Fail Failure type if admission controller is removed", func() {
		proxyType = createProxyType()
		webhook = createNamespacedWebhook(admissionv1beta1.Fail)

		Expect(kubeClient.Delete(context.TODO(), admDeploy)).To(Succeed())
		admDeploy = nil
		Eventually(func () error { return common.VerifyNoEndpoint(admService.Name, admService.Namespace) }, 60, 5).Should(Succeed())

		pod = getPod(false)
		tryCreatePod(&pod, false, false)
	})
})


func createProxyType() *v1alpha1.ProxyValidatingType {
	By("Add ProxyValidatingType")
	pt := &v1alpha1.ProxyValidatingType{
		ObjectMeta: metav1.ObjectMeta{
			Name: "gesher-test",
		},
		Spec: v1alpha1.ProxyValidatingTypeSpec{
			Types: []admissionv1beta1.RuleWithOperations{{
				Operations: []admissionv1beta1.OperationType{"CREATE"},
				Rule: admissionv1beta1.Rule{
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
			Name: fmt.Sprintf("pod-test-%v", time.Now().Unix()),
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

func createNamespacedWebhook(failurePolicy admissionv1beta1.FailurePolicyType) *v1alpha1.NamespacedValidatingProxy {
	By("Add NamespacedValidatingProxy")
	path := "/admission"

	nvp := &v1alpha1.NamespacedValidatingProxy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-hook",
			Namespace: common.Namespace,
		},
		Spec: v1alpha1.NamespacedValidatingProxySpec{
			Webhooks: []admissionv1beta1.ValidatingWebhook{
				{
					Name:                    "test-hook",
					FailurePolicy:			 &failurePolicy,
					ClientConfig:            admissionv1beta1.WebhookClientConfig{
						Service:  &admissionv1beta1.ServiceReference{
							Name: admService.Name,
							Namespace: admService.Namespace,
							Port: &admService.Spec.Ports[0].Port,
							Path: &path,
						},
						CABundle: admSecret.Data["cert"],
					},
					Rules:                   []admissionv1beta1.RuleWithOperations{
						{
							Operations: []admissionv1beta1.OperationType{admissionv1beta1.Create},
							Rule:       admissionv1beta1.Rule{
								APIGroups:   []string{""},
								APIVersions: []string{"v1"},
								Resources:   []string{"pods"},
							},
						},
					},
				},
			},
		},
	}

	Expect(kubeClient.Create(context.TODO(), nvp)).To(Succeed())

	By( "wait on resource to be applied")
	Eventually(func() error { return common.VerifyApplied(nvp) }, 60, 5).Should(Succeed())

	return nvp
}

func deleteNamespacedWebhook(nvp *v1alpha1.NamespacedValidatingProxy) *v1alpha1.NamespacedValidatingProxy {
	Expect(kubeClient.Delete(context.TODO(), nvp)).To(Succeed())
	Eventually(func() error { return common.VerifyDeleted(nvp) }, 60, 5).Should(Succeed())

	return nil
}
