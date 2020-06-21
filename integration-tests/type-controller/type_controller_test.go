package type_controller_test

import (
	"context"
	"github.com/redislabs/gesher/pkg/apis/app/v1alpha1"
	admissionv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	namePrefix   = "proxy-validating-type"
	testGroup    = "TestGroup"
	testVersion  = "TestVersion"
	testResource = "TestResources"
	op1          = admissionv1beta1.Create
	op2          = admissionv1beta1.Delete
)

var _ = Describe("TypeController", func() {
	var (
		pt  *v1alpha1.ProxyValidatingType
	)
	BeforeEach(func() {
		pt = &v1alpha1.ProxyValidatingType{
			ObjectMeta: metav1.ObjectMeta{
				Name: namePrefix,
			},
			Spec:   v1alpha1.ProxyValidatingTypeSpec{
				Types: []admissionv1beta1.RuleWithOperations{{
					Operations: []admissionv1beta1.OperationType{op1},
					Rule:       admissionv1beta1.Rule{
						APIGroups:   []string{testGroup},
						APIVersions: []string{testVersion},
						Resources:   []string{testResource},
					},
				}},
			},
		}
	})

	AfterEach(func() {
		Expect(kubeClient.DeleteAllOf(context.TODO(), &v1alpha1.ProxyValidatingType{})).To(Succeed())
		Expect(verifyEmpty()).To(Succeed())
	})

	It("Adding a Single Custom Resource", func() {
		pt1 := pt.DeepCopy(); pt1.Name = "proxy-validating-type-1"

		By("Add resource 1")
		Expect(kubeClient.Create(context.TODO(), pt1)).To(Succeed())

		By("wait on resource")
		Expect(verifyApplied(pt1)).To(Succeed())

		By("validate webhook")
		Expect(validateInWebhook([]*v1alpha1.ProxyValidatingType{pt1})).To(Succeed())
	})

	It("Adding Multiple Custom Resource", func() {
		pt1 := pt.DeepCopy(); pt1.Name = "proxy-validating-type-1"
		pt2 := pt.DeepCopy(); pt1.Name = "proxy-validating-type-2"

		By("Add resource 1")
		Expect(kubeClient.Create(context.TODO(), pt1)).To(Succeed())

		By("Add resource 2")
		Expect(kubeClient.Create(context.TODO(), pt2)).To(Succeed())

		By("wait on resources")
		Expect(verifyApplied(pt1)).To(Succeed())
		Expect(verifyApplied(pt2)).To(Succeed())

		By("validate webhook")
		Expect(validateInWebhook([]*v1alpha1.ProxyValidatingType{pt1, pt2})).To(Succeed())
	})

	It("Modifying a Single Custom Resource", func() {
		pt1 := pt.DeepCopy(); pt1.Name = "proxy-validating-type-1"

		By("Add resouce 1")
		Expect(kubeClient.Create(context.TODO(), pt1)).To(Succeed())

		By("wait on resources")
		Expect(verifyApplied(pt1)).To(Succeed())

		By("Update resouce 1 to 1a")
		pt1.Spec.Types[0].Operations[0] = op2
		Expect(kubeClient.Update(context.TODO(), pt1)).To(Succeed())

		By("validate webhook")
		Expect(validateInWebhook([]*v1alpha1.ProxyValidatingType{pt1})).To(Succeed())
	})

	It("Adding a Duplicate Custom Resource", func() {
		pt1 := pt.DeepCopy(); pt1.Name = "proxy-validating-type-1"
		pt2 := pt.DeepCopy(); pt1.Name = "proxy-validating-type-2"

		By("Add resouce 1")
		Expect(kubeClient.Create(context.TODO(), pt1)).To(Succeed())

		By("Add resource 2")
		Expect(kubeClient.Create(context.TODO(), pt2)).To(Succeed())

		By("wait on resources")
		Expect(verifyApplied(pt1)).To(Succeed())
		Expect(verifyApplied(pt2)).To(Succeed())

		By("Validate webhook")
		Expect(validateInWebhook([]*v1alpha1.ProxyValidatingType{pt1})).To(Succeed())
		Expect(validateInWebhook([]*v1alpha1.ProxyValidatingType{pt2})).To(Succeed())

		By("Delete resouce 1")
		Expect(kubeClient.Delete(context.TODO(), pt1)).To(Succeed())
		Expect(verifyDeleted(pt1)).To(Succeed())

		By("validate webhook")
		Expect(validateInWebhook([]*v1alpha1.ProxyValidatingType{pt2})).To(Succeed())
	})

	It("Adding a Similar Custom Resource", func() {
		By("Add resouce 1")	
		pt1 := pt.DeepCopy(); pt1.Name = "proxy-validating-type-1"
		Expect(kubeClient.Create(context.TODO(), pt1)).To(Succeed())

		By("Add resource 1a")
		pt1a := pt.DeepCopy(); pt1a.Name = "proxy-validating-type-2" ; pt1a.Spec.Types[0].Operations[0] = op2
		Expect(kubeClient.Create(context.TODO(), pt1a)).To(Succeed())

		By("wait on resources")
		Expect(verifyApplied(pt1)).To(Succeed())
		Expect(verifyApplied(pt1a)).To(Succeed())

		By("validate webhook")
		Expect(validateInWebhook([]*v1alpha1.ProxyValidatingType{pt1})).To(Succeed())
		Expect(validateInWebhook([]*v1alpha1.ProxyValidatingType{pt1a})).To(Succeed())
	})

	It("Deleting a Similiar Custom Resource", func() {
		By("Add resouce 1")
		pt1 := pt.DeepCopy(); pt1.Name = "proxy-validating-type-1"
		Expect(kubeClient.Create(context.TODO(), pt1)).To(Succeed())

		By("Add resource 1a")
		pt1a := pt.DeepCopy(); pt1a.Name = "proxy-validating-type-2" ; pt1a.Spec.Types[0].Operations[0] = op2
		Expect(kubeClient.Create(context.TODO(), pt1a)).To(Succeed())

		By("wait on resources")
		Expect(verifyApplied(pt1)).To(Succeed())
		Expect(verifyApplied(pt1a)).To(Succeed())

		By("Delete resouce 1")
		Expect(kubeClient.Delete(context.TODO(), pt1)).To(Succeed())
		Expect(verifyDeleted(pt1)).To(Succeed())

		By("validate webhook")
		Expect(validateInWebhook([]*v1alpha1.ProxyValidatingType{pt1a})).To(Succeed())
		Expect(validateNotInWebhook([]*v1alpha1.ProxyValidatingType{pt1})).To(Succeed())
	})
})
