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

	"github.com/redislabs/gesher/integration-tests/common"
	"github.com/redislabs/gesher/pkg/apis/app/v1alpha1"
	admregv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	namePrefix   = "namespaced-validating-type"
	testGroup    = "TestGroup"
	testVersion  = "TestVersion"
	testResource = "TestResources"
	op1          = admregv1.Create
	op2          = admregv1.Delete
)

var _ = Describe("TypeController", func() {
	var (
		pt *v1alpha1.NamespacedValidatingType
	)
	BeforeEach(func() {
		pt = &v1alpha1.NamespacedValidatingType{
			ObjectMeta: metav1.ObjectMeta{
				Name: namePrefix,
			},
			Spec: v1alpha1.NamespacedValidatingTypeSpec{
				Types: []admregv1.RuleWithOperations{{
					Operations: []admregv1.OperationType{op1},
					Rule: admregv1.Rule{
						APIGroups:   []string{testGroup},
						APIVersions: []string{testVersion},
						Resources:   []string{testResource},
					},
				}},
			},
		}
	})

	AfterEach(func() {
		Expect(kubeClient.DeleteAllOf(context.TODO(), &v1alpha1.NamespacedValidatingType{})).To(Succeed())
		Eventually(common.VerifyEmpty, 60, 5).Should(Succeed())
	})

	It("Adding a Single Custom Resource", func() {
		pt1 := pt.DeepCopy()
		pt1.Name = "namespaced-validating-type-1"

		By("Add resource 1")
		Expect(kubeClient.Create(context.TODO(), pt1)).To(Succeed())

		By("wait on resource")
		Eventually(func() error { return common.VerifyApplied(pt1) }, 60, 5).Should(Succeed())

		By("validate webhook")
		Expect(common.ValidateInWebhook([]*v1alpha1.NamespacedValidatingType{pt1})).To(Succeed())
	})

	It("Adding Multiple Custom Resource", func() {
		pt1 := pt.DeepCopy()
		pt1.Name = "namespaced-validating-type-1"
		pt2 := pt.DeepCopy()
		pt1.Name = "namespaced-validating-type-2"

		By("Add resource 1")
		Expect(kubeClient.Create(context.TODO(), pt1)).To(Succeed())

		By("Add resource 2")
		Expect(kubeClient.Create(context.TODO(), pt2)).To(Succeed())

		By("wait on resources")
		Eventually(func() error { return common.VerifyApplied(pt1) }, 60, 5).Should(Succeed())
		Eventually(func() error { return common.VerifyApplied(pt2) }, 60, 5).Should(Succeed())

		By("validate webhook")
		Expect(common.ValidateInWebhook([]*v1alpha1.NamespacedValidatingType{pt1, pt2})).To(Succeed())
	})

	It("Modifying a Single Custom Resource", func() {
		pt1 := pt.DeepCopy()
		pt1.Name = "namespaced-validating-type-1"

		By("Add resouce 1")
		Expect(kubeClient.Create(context.TODO(), pt1)).To(Succeed())

		By("wait on resources")
		Eventually(func() error { return common.VerifyApplied(pt1) }, 60, 5).Should(Succeed())

		By("Update resouce 1 to 1a")
		pt1.Spec.Types[0].Operations[0] = op2
		Expect(kubeClient.Update(context.TODO(), pt1)).To(Succeed())

		By("Wait on Update")
		Eventually(func() error { return common.VerifyApplied(pt1) }, 60, 5).Should(Succeed())

		By("validate webhook")
		Expect(common.ValidateInWebhook([]*v1alpha1.NamespacedValidatingType{pt1})).To(Succeed())
	})

	It("Adding a Duplicate Custom Resource", func() {
		pt1 := pt.DeepCopy()
		pt1.Name = "namespaced-validating-type-1"
		pt2 := pt.DeepCopy()
		pt1.Name = "namespaced-validating-type-2"

		By("Add resouce 1")
		Expect(kubeClient.Create(context.TODO(), pt1)).To(Succeed())

		By("Add resource 2")
		Expect(kubeClient.Create(context.TODO(), pt2)).To(Succeed())

		By("wait on resources")
		Eventually(func() error { return common.VerifyApplied(pt1) }, 60, 5).Should(Succeed())
		Eventually(func() error { return common.VerifyApplied(pt2) }, 60, 5).Should(Succeed())

		By("Validate webhook")
		Expect(common.ValidateInWebhook([]*v1alpha1.NamespacedValidatingType{pt1})).To(Succeed())
		Expect(common.ValidateInWebhook([]*v1alpha1.NamespacedValidatingType{pt2})).To(Succeed())

		By("Delete resouce 1")
		Expect(kubeClient.Delete(context.TODO(), pt1)).To(Succeed())
		Eventually(func() error { return common.VerifyDeleted(pt1) }, 60, 5).Should(Succeed())

		By("validate webhook")
		Expect(common.ValidateInWebhook([]*v1alpha1.NamespacedValidatingType{pt2})).To(Succeed())
	})

	It("Adding a Similar Custom Resource", func() {
		By("Add resouce 1")
		pt1 := pt.DeepCopy()
		pt1.Name = "namespaced-validating-type-1"
		Expect(kubeClient.Create(context.TODO(), pt1)).To(Succeed())

		By("Add resource 1a")
		pt1a := pt.DeepCopy()
		pt1a.Name = "namespaced-validating-type-2"
		pt1a.Spec.Types[0].Operations[0] = op2
		Expect(kubeClient.Create(context.TODO(), pt1a)).To(Succeed())

		By("wait on resources")
		Eventually(func() error { return common.VerifyApplied(pt1) }, 60, 5).Should(Succeed())
		Eventually(func() error { return common.VerifyApplied(pt1a) }, 60, 5).Should(Succeed())

		By("validate webhook")
		Expect(common.ValidateInWebhook([]*v1alpha1.NamespacedValidatingType{pt1})).To(Succeed())
		Expect(common.ValidateInWebhook([]*v1alpha1.NamespacedValidatingType{pt1a})).To(Succeed())
	})

	It("Deleting a Similiar Custom Resource", func() {
		By("Add resouce 1")
		pt1 := pt.DeepCopy()
		pt1.Name = "namespaced-validating-type-1"
		Expect(kubeClient.Create(context.TODO(), pt1)).To(Succeed())

		By("Add resource 1a")
		pt1a := pt.DeepCopy()
		pt1a.Name = "namespaced-validating-type-2"
		pt1a.Spec.Types[0].Operations[0] = op2
		Expect(kubeClient.Create(context.TODO(), pt1a)).To(Succeed())

		By("wait on resources")
		Eventually(func() error { return common.VerifyApplied(pt1) }, 60, 5).Should(Succeed())
		Eventually(func() error { return common.VerifyApplied(pt1a) }, 60, 5).Should(Succeed())

		By("Delete resouce 1")
		Expect(kubeClient.Delete(context.TODO(), pt1)).To(Succeed())
		Eventually(func() error { return common.VerifyDeleted(pt1) }, 60, 5).Should(Succeed())

		By("validate webhook")
		Expect(common.ValidateInWebhook([]*v1alpha1.NamespacedValidatingType{pt1a})).To(Succeed())
		Expect(common.ValidateNotInWebhook([]*v1alpha1.NamespacedValidatingType{pt1})).To(Succeed())
	})
})
