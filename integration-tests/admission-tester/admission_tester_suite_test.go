package admission_tester_test

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"

	"github.com/redislabs/gesher/integration-tests/common"
	"k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
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
	sa          *corev1.ServiceAccount
	role        *rbacv1beta1.ClusterRole
	roleBinding *rbacv1beta1.ClusterRoleBinding

	webhook *v1beta1.ValidatingWebhookConfiguration
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
	role = common.LoadClusterRole()
	roleBinding = common.LoadClusterRoleBinding()

	service = loadService()
	deploy = loadDeploy()

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
	if role != nil {
		Expect(kubeClient.Delete(context.TODO(), role)).To(Succeed())
		role = nil
	}
	if roleBinding != nil {
		Expect(kubeClient.Delete(context.TODO(), roleBinding)).To(Succeed())
		roleBinding = nil
	}
})

func loadService() *corev1.Service {
	By("Read and load Service")

	s := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "admission-test",
			Namespace: common.Namespace,
		},
		Spec:       corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Protocol:   "TCP",
				Port:       443,
				TargetPort: intstr.IntOrString{IntVal: 9443},
			}},
			Selector: map[string]string{"app": "admission-test"},
		},
	}

	Expect(kubeClient.Create(context.TODO(), s)).To(Succeed())

	return s
}

func loadDeploy() *appsv1.Deployment {
	By("Read and Load Deployment")

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "admission-test",
			Namespace: common.Namespace,
		},
		Spec:       appsv1.DeploymentSpec{
			Selector:                &metav1.LabelSelector{
				MatchLabels:      map[string]string{"app": "admission-test"},
			},
			Template:                corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "admission-test"},
				},
				Spec:       corev1.PodSpec{
					ServiceAccountName: "gesher",
					Containers:                    []corev1.Container{
						{
							Name: "admission-test",
							Image: "quay.io/spotter/gesher-admisison-test:test",
							Command: []string{"/admission"},
							Env: []corev1.EnvVar{
								{
									Name: "POD_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef:         &corev1.ObjectFieldSelector{
											FieldPath:  "metadata.namespace",
										},
									},
								},
							},
							ImagePullPolicy: corev1.PullAlways,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 9443,
								},
							},
							ReadinessProbe: &corev1.Probe{
								Handler:             corev1.Handler{
									HTTPGet:   &corev1.HTTPGetAction{
										Path:        "/healthz",
										Port:        intstr.IntOrString{IntVal: 9443},
										Scheme:      "HTTPS",
									},
								},
								TimeoutSeconds:      5,
								PeriodSeconds:       5,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
						},
					},

				},
			},
		},
	}

	Expect(kubeClient.Create(context.TODO(), deploy)).To(Succeed())

	Eventually(func() error { return common.WaitForDeployment(deploy) }, 60, 5).Should(Succeed())

	return deploy
}

func loadWebhook(s *corev1.Secret) *v1beta1.ValidatingWebhookConfiguration {
	By("Read and Load Webhook")
	
	path := "/admission"
	failurePolicy := v1beta1.Fail
	scope         := v1beta1.AllScopes

	webhook := &v1beta1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:                       "admission-test",
		},
		Webhooks:   []v1beta1.ValidatingWebhook{
			{
				Name:                    "test.admission.gesher",
				ClientConfig:            v1beta1.WebhookClientConfig{
					Service:  &v1beta1.ServiceReference{
						Name: "admission-test",
						Namespace: common.Namespace,
						Path: &path, 
					},
					CABundle: s.Data["cert"],
				},
				Rules:                   []v1beta1.RuleWithOperations{
					{
						Operations: []v1beta1.OperationType{v1beta1.Create},
						Rule:       v1beta1.Rule{
							APIGroups:   []string{""},
							APIVersions: []string{"v1"},
							Resources:   []string{"namespaces"},
							Scope: &scope,
						},
					},
				},
				FailurePolicy: &failurePolicy,
			},
		},
	}

	Expect(kubeClient.Create(context.TODO(), webhook)).To(Succeed())

	return webhook
}
