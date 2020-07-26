package https_server_test

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
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

	crd1        *apiextv1beta1.CustomResourceDefinition
	crd2        *apiextv1beta1.CustomResourceDefinition
	service     *corev1.Service
	sa          *corev1.ServiceAccount
	role        *rbacv1beta1.ClusterRole
	roleBinding *rbacv1beta1.ClusterRoleBinding
)

var _ = BeforeSuite(func() {
	var err error

	By("Get client")
	kubeClient, _, err = common.GetClient()
	Expect(err).To(Succeed())

	crd1 = common.LoadProxyValidatingTypeCRD()
	crd2 = common.LoadNamespacedValidatingProxyCRD()
	service = common.LoadService()
	serviceName = service.Name
	sa = common.LoadServiceAccount()
	role = common.LoadClusterRole()
	roleBinding = common.LoadClusterRoleBinding()
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
	if role != nil {
		Expect(kubeClient.Delete(context.TODO(), role)).To(Succeed())
		role = nil
	}
	if roleBinding != nil {
		Expect(kubeClient.Delete(context.TODO(), roleBinding)).To(Succeed())
		roleBinding = nil
	}

})
