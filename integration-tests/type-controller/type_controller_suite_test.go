package type_controller_test

import (
	"context"
	"github.com/redislabs/gesher/cmd/manager/flags"
	"github.com/redislabs/gesher/integration-tests/common"
	corev1 "k8s.io/api/core/v1"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTypeController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TypeController Suite")
}

var (
	crd         *v1beta1.CustomResourceDefinition
	deploy      *appsv1.Deployment
	sa          *corev1.ServiceAccount
	service     *corev1.Service
	role        *rbacv1beta1.ClusterRole
	roleBinding *rbacv1beta1.ClusterRoleBinding
	secret      *corev1.Secret

	kubeClient client.Client
)

var _ = BeforeSuite(func() {
	var err error

	By("Setup kube clients")
	kubeClient, _, err = common.GetClient()
	Expect(err).To(Succeed())

	crd = common.LoadCRD()
	service = common.LoadService()
	sa = common.LoadServiceAccount()
	role = common.LoadClusterRole()
	roleBinding = common.LoadClusterRoleBinding()
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
	if crd != nil {
		Expect(kubeClient.Delete(context.TODO(), crd)).To(Succeed())
		crd = nil
	}

	if service != nil {
		Expect(kubeClient.Delete(context.TODO(), service)).To(Succeed())
		service = nil
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

	if secret != nil {
		Expect(kubeClient.Delete(context.TODO(), secret)).To(Succeed())
		secret = nil
	}
})
