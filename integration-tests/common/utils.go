package common

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/redislabs/gesher/cmd/manager/flags"
	"io/ioutil"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/googleapis/gnostic/compiler"

	"github.com/redislabs/gesher/pkg/controller/proxyvalidatingtype"
	"k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/redislabs/gesher/pkg/apis"
	"github.com/redislabs/gesher/pkg/apis/app/v1alpha1"

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	webhookResourceName = proxyvalidatingtype.ProxyWebhookName
)

var (
	kubeClient client.Client
	cl         kubernetes.Interface
	Namespace  = flags.DefaultNamespace // make configurable
)

func GetClient() (client.Client, kubernetes.Interface, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, nil, err
	}

	mapper, err := apiutil.NewDynamicRESTMapper(cfg)
	if err != nil {
		return nil, nil, err
	}

	err = apis.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, nil, err
	}

	err = apiext.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, nil, err
	}

	kubeClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme, Mapper: mapper})
	if err != nil {
		return nil, nil, err
	}
	cl, err = kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, nil, err
	}

	return kubeClient, cl, nil
}

func LoadCRD() *apiextv1beta1.CustomResourceDefinition {
	By("Read and Load CRD")

	c := &apiextv1beta1.CustomResourceDefinition{}

	data, err := ioutil.ReadFile("../../deploy/crds/app.redislabs.com_proxyvalidatingtypes_crd.yaml")
	Expect(err).To(BeNil())
	Expect(yaml.NewYAMLToJSONDecoder(bytes.NewReader(data)).Decode(c)).To(Succeed())
	Expect(kubeClient.Create(context.TODO(), c)).To(Succeed())

	return c
}

func LoadServiceAccount() *v1.ServiceAccount {
	By("Read and Load ServiceAccount")

	sa := &v1.ServiceAccount{}

	data, err := ioutil.ReadFile("../../deploy/service_account.yaml")
	Expect(err).To(BeNil())
	Expect(yaml.NewYAMLToJSONDecoder(bytes.NewReader(data)).Decode(sa)).To(Succeed())
	sa.Namespace = Namespace
	Expect(kubeClient.Create(context.TODO(), sa)).To(Succeed())

	return sa
}

func LoadClusterRole() *rbacv1beta1.ClusterRole {
	By("Read and Load Role")

	role := &rbacv1beta1.ClusterRole{}

	data, err := ioutil.ReadFile("../../deploy/role.yaml")
	Expect(err).To(BeNil())
	Expect(yaml.NewYAMLToJSONDecoder(bytes.NewReader(data)).Decode(role)).To(Succeed())
	Expect(kubeClient.Create(context.TODO(), role)).To(Succeed())

	return role
}

func LoadClusterRoleBinding() *rbacv1beta1.ClusterRoleBinding {
	By("Read and Load RoleBinding")

	roleBinding := &rbacv1beta1.ClusterRoleBinding{}

	data, err := ioutil.ReadFile("../../deploy/role_binding.yaml")
	Expect(err).To(BeNil())
	Expect(yaml.NewYAMLToJSONDecoder(bytes.NewReader(data)).Decode(roleBinding)).To(Succeed())
	Expect(kubeClient.Create(context.TODO(), roleBinding)).To(Succeed())

	return roleBinding
}

func LoadService() *v1.Service {
	By("Read and load Service")

	s := &v1.Service{}

	data, err := ioutil.ReadFile("../../deploy/service.yaml")
	Expect(err).To(BeNil())
	Expect(yaml.NewYAMLToJSONDecoder(bytes.NewReader(data)).Decode(s)).To(Succeed())
	s.Namespace = Namespace
	Expect(kubeClient.Create(context.TODO(), s)).To(Succeed())

	return s
}

func LoadOperator(desc string) *appsv1.Deployment {
	By(desc)

	deploy := &appsv1.Deployment{}

	data, err := ioutil.ReadFile("../../deploy/operator.yaml")
	Expect(err).To(BeNil())
	Expect(yaml.NewYAMLToJSONDecoder(bytes.NewReader(data)).Decode(deploy)).To(Succeed())
	deploy.Namespace = Namespace
	Expect(kubeClient.Create(context.TODO(), deploy)).To(Succeed())

	Eventually(func() error { return WaitForDeployment(deploy) }, 60, 5).Should(Succeed())

	return deploy
}

func WaitForDeployment(deploy *appsv1.Deployment) error {
	fmt.Fprintf(GinkgoWriter, "Verifying that new deployment is ready: ")
	d, err := cl.AppsV1().Deployments(deploy.Namespace).Get(context.TODO(), deploy.Name, metav1.GetOptions{})

	// commented out as wasn't populating status, but would prefer to use if can be made to work
	//err := kubeClient.Get(context.TODO(), types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, d)
	if err == nil {
		if d.Status.UpdatedReplicas == *d.Spec.Replicas && d.Status.ReadyReplicas == *d.Spec.Replicas {
			fmt.Fprintf(GinkgoWriter, "Ready\n")
			return nil
		}
	} else if !apierrors.IsNotFound(err) {
		fmt.Fprintf(GinkgoWriter, "Error: %v\n", err)
		return err
	}

	err = fmt.Errorf("%v not available yet: %+v", deploy.Name, deploy.Status)
	fmt.Fprintf(GinkgoWriter, "%v\n", err)
	return err
}

func VerifyEmpty() error {
	fmt.Fprintf(GinkgoWriter, "Verifying that Webhook %v is Empty: ", webhookResourceName)

	item := &v1beta1.ValidatingWebhookConfiguration{}
	err := kubeClient.Get(context.TODO(), client.ObjectKey{Name: webhookResourceName}, item)
	if err != nil {
		fmt.Fprintf(GinkgoWriter, "%v\n", err)
		return err
	}

	switch len(item.Webhooks) {
	case 0:
		fmt.Fprintf(GinkgoWriter, "Success!\n")
		return nil
	case 1:
		if len(item.Webhooks[0].Rules) == 0 {
			fmt.Fprintf(GinkgoWriter, "Success!\n")
			return nil
		}
		err := errors.New("expected no Rules")
		fmt.Fprintf(GinkgoWriter, "Failed: %v\n", err)
		return err
	default:
		err := errors.New("more than one webhook specified")
		fmt.Fprintf(GinkgoWriter, "Failed: %v\n", err)
		return err
	}
}

func VerifyApplied(pt *v1alpha1.ProxyValidatingType) error {
	fmt.Fprintf(GinkgoWriter, "Verifying that PVR %v was applied by operator: ", pt.Name)

	prevGen := pt.Status.ObservedGeneration
	name := pt.Name

	err := kubeClient.Get(context.TODO(), client.ObjectKey{Name: name}, pt)
	if err != nil {
		fmt.Fprintf(GinkgoWriter, "Failed: %v\n", err)
		return err
	}
	if prevGen == pt.Status.ObservedGeneration || pt.Generation != pt.Status.ObservedGeneration {
		err := errors.New("operator hasn't updated generation in status yet")
		fmt.Fprintf(GinkgoWriter, "Failed: %v\n", err)
		return err
	}

	fmt.Fprintf(GinkgoWriter, "Success!\n")
	return nil
}

func VerifyDeleted(pt *v1alpha1.ProxyValidatingType) error {
	fmt.Fprintf(GinkgoWriter, "Verifying that PVR %v was deleted: ", pt.Name)

	name := pt.Name

	err := kubeClient.Get(context.TODO(), client.ObjectKey{Name: name}, pt)
	if err == nil {
		err := fmt.Errorf("%v not deleted yet", pt.Name)
		fmt.Fprintf(GinkgoWriter, "Failed: %v\n", err)
		return err
	}
	if !apierrors.IsNotFound(err) {
		fmt.Fprintf(GinkgoWriter, "Failed: %v\n", err)
		return err
	}

	fmt.Fprintf(GinkgoWriter, "Success!\n")
	return nil
}

func ValidateInWebhook(ptList []*v1alpha1.ProxyValidatingType) error {
	item := &v1beta1.ValidatingWebhookConfiguration{}
	err := kubeClient.Get(context.TODO(), client.ObjectKey{Name: webhookResourceName}, item)
	if err != nil {
		return err
	}

	if len(item.Webhooks) != 1 {
		return errors.New("expected only a single webhook")
	}

	for _, pt := range ptList {
		if !proxyValidatingTypeExists(pt, item.Webhooks[0].Rules) {
			return fmt.Errorf("couldn't validate %+v in %+v", pt, item.Webhooks[0].Rules)
		}
	}

	return nil
}

func ValidateNotInWebhook(ptList []*v1alpha1.ProxyValidatingType) error {
	item := &v1beta1.ValidatingWebhookConfiguration{}
	err := kubeClient.Get(context.TODO(), client.ObjectKey{Name: webhookResourceName}, item)
	if err != nil {
		return err
	}

	if len(item.Webhooks) != 1 {
		return errors.New("expected only a single webhook")
	}

	for _, pt := range ptList {
		if proxyValidatingTypeExists(pt, item.Webhooks[0].Rules) {
			return fmt.Errorf("%+v still exists in %+v", pt, item.Webhooks[0].Rules)
		}
	}

	return nil
}

func proxyValidatingTypeExists(pt *v1alpha1.ProxyValidatingType, rules []v1beta1.RuleWithOperations) bool {
	for _, pType := range pt.Spec.Types {
		for _, group := range pType.APIGroups {
			for _, version := range pType.APIVersions {
				for _, resource := range pType.Resources {
				loop:
					for _, op := range pType.Operations {
						for _, rule := range rules {
							if compiler.StringArrayContainsValue(rule.APIGroups, group) &&
								compiler.StringArrayContainsValue(rule.APIVersions, version) &&
								compiler.StringArrayContainsValue(rule.Resources, resource) &&
								OpArrayContainsValues(rule.Operations, op) {
								// found a match, don't have to check anymore webhook rules, continue checking pType
								continue loop
							}
						}
						// only way to hit this should be if exhaust rules without finding a match
						return false
					}
				}
			}
		}
	}

	return true
}

func OpArrayContainsValues(operations []v1beta1.OperationType, op v1beta1.OperationType) bool {
	for _, operation := range operations {
		if op == operation {
			return true
		}
	}

	return false
}

func VerifyPodSuccess(p *v1.Pod) error {
	fmt.Fprintf(GinkgoWriter, "Verifying Pod %v Completes Succesfully: ", p.Name)

	err := kubeClient.Get(context.TODO(), client.ObjectKey{Name: p.Name, Namespace: p.Namespace}, p)
	if err != nil {
		fmt.Fprintf(GinkgoWriter, "Failed: %v\n", err)
		return err
	}

	if p.Status.Phase == v1.PodRunning || p.Status.Phase == v1.PodPending {
		err := errors.New("Pod Is Still Running")
		fmt.Fprintf(GinkgoWriter, "Failed: %v\n", err)
		return err
	}

	if p.Status.Phase != v1.PodSucceeded || p.Status.ContainerStatuses[0].State.Terminated.ExitCode != 0 {
		err := fmt.Errorf("%v failed", p.Name)
		fmt.Fprintf(GinkgoWriter, "Failed: %v\n", err)
		return err
	}

	fmt.Fprintf(GinkgoWriter, "Success!\n")
	return nil
}

func VerifyEndpoint(e, namespace string) error {
	fmt.Fprintf(GinkgoWriter, "Verifying Endpoint: ")

	var endpoint v1.Endpoints

	err := kubeClient.Get(context.TODO(), types.NamespacedName{Name: e, Namespace: namespace}, &endpoint)
	if err != nil {
		fmt.Fprintf(GinkgoWriter, "Failed: %v\n", err)
		return err
	}

	if len(endpoint.Subsets) == 0 {
		err := errors.New("endpoint not filled in yet")
		fmt.Fprintf(GinkgoWriter, "Failed: %v\n", err)
		return err
	}

	if len(endpoint.Subsets[0].Addresses) == 0 {
		err := errors.New("endpoint addreses not filled in yet")
		fmt.Fprintf(GinkgoWriter, "Failed: %v\n", err)
		return err
	}

	fmt.Fprintf(GinkgoWriter, "Success!\n")
	return nil
}
