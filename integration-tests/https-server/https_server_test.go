package https_server_test

import (
	"context"
	"fmt"
	"path/filepath"
	"reflect"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/redislabs/gesher/cmd/manager/flags"
	"github.com/redislabs/gesher/integration-tests/common"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HttpsServer", func() {
	var (
		deploy *appsv1.Deployment
		secret *v1.Secret
		pods   []*v1.Pod
	)

	// Treating BeforeEach as invariant enforcement
	BeforeEach(func() {
		var s v1.Secret
		var d appsv1.Deployment

		Expect(kubeClient.Get(context.TODO(), types.NamespacedName{Name: flags.DefaultTlsSecret, Namespace: common.Namespace}, &s)).ToNot(Succeed())
		Expect(kubeClient.Get(context.TODO(), types.NamespacedName{Name: "gesher", Namespace: common.Namespace}, &d)).ToNot(Succeed())
	})

	AfterEach(func() {
		// Delete deployment
		if deploy != nil {
			Expect(kubeClient.Delete(context.TODO(), deploy)).To(Succeed())
			deploy = nil
		}
		if secret != nil {
			Expect(kubeClient.Delete(context.TODO(), secret)).To(Succeed())
			secret = nil
		}
		if len(pods) != 0 {
			for _, pod := range pods {
				Expect(kubeClient.Delete(context.TODO(), pod)).To(Succeed())
			}
			pods = nil
		}
	})

	It("TLS Testing", func() {
		var s1, s2 v1.Secret

		deploy = common.LoadOperator("Read and Load Operator with Secret Creation")

		By("Verify that the secret can be read and its keys")

		Expect(kubeClient.Get(context.TODO(), types.NamespacedName{Name: flags.DefaultTlsSecret, Namespace: common.Namespace}, &s1)).To(Succeed())
		secret = &s1
		Expect(s1.Data).To(HaveKey("cert"))
		Expect(s1.Data).To(HaveKey("privateKey"))

		By("Run Curl HTTPS Test")
		runCurl(&pods)

		By("Delete Operator Deployment")
		Expect(kubeClient.Delete(context.TODO(), deploy)).To(Succeed())
		deploy = nil

		deploy = common.LoadOperator("Redeploy Operator to Reuse Existing Secret")

		By("Get and Compare Secret")
		Expect(kubeClient.Get(context.TODO(), types.NamespacedName{Name: flags.DefaultTlsSecret, Namespace: common.Namespace}, &s2)).To(Succeed())
		Expect(reflect.DeepEqual(s1, s2)).To(BeTrue())

		By("Run Curl HTTPS Test Again")
		runCurl(&pods)
	})
})

func runCurl(pods *[]*v1.Pod) {
	Eventually(func() error { return common.VerifyEndpoint("gesher", common.Namespace) }, 60, 5).Should(Succeed())
	p := createCurlPod()
	Expect(kubeClient.Create(context.TODO(), p)).To(Succeed())
	*pods = append(*pods, p)
	Eventually(func() error { return common.VerifyPodSuccess(p) }, 60, 5).Should(Succeed())
}

func createCurlPod() *v1.Pod {
	url := fmt.Sprintf("https://%v.%v/healthz", service.Name, common.Namespace)
	cacrtDir := "/cacrt"
	cacrt := filepath.Join(cacrtDir, "cert")
	image := "curlimages/curl"

	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("https-test-%v", time.Now().Unix()),
			Namespace: common.Namespace,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Image: image,
					Name:  "https-test",
					Args: []string{
						"curl",
						"-vvv",
						"--connect-timeout", "5",
						"--max-time", "60",
						"--retry", "30",
						"--retry-delay", "5",
						"--retry-max-time", "120",
						"--retry-connrefused",
						"--cacert", cacrt,
						url,
					},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "tls",
							MountPath: cacrtDir,
						},
					},
				},
			},
			Volumes: []v1.Volume{
				{
					Name: "tls",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: flags.DefaultTlsSecret,
						},
					},
				},
			},
			RestartPolicy: v1.RestartPolicyNever,
		},
	}
}
