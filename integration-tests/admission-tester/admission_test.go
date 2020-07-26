package admission_tester_test

import (
	"context"
	admissionTest "github.com/redislabs/gesher/pkg/admission-test"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Admission", func() {
	var (
		ns *corev1.Namespace
	)

	AfterEach(func() {
		if ns != nil {
			Expect(kubeClient.Delete(context.TODO(), ns)).To(Succeed())
			ns = nil
		}
	})

	It("Test Denied", func() {
		ns = getNamespace("denied")

		By("Try to create a namespace without the proper label")
		Expect(kubeClient.Create(context.TODO(), ns)).To(MatchError(MatchRegexp("admission-allow label not set to allow this resource")))
		ns = nil
	})

	It("Test Allow", func() {
		ns = getNamespace("allowed")
		ns.Labels = map[string]string{admissionTest.AdmissionKey: admissionTest.AdmissionAllow}

		By("Try to create a namespace with the proper label")
		Expect(kubeClient.Create(context.TODO(), ns)).To(Succeed())
	})
})

func getNamespace(name string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Labels:    nil,
		},
	}
}
