package type_controller_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"


	v1 "k8s.io/api/apps/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTypeController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TypeController Suite")
}

var (
	crd *v1beta1.CustomResourceDefinition
	deploy *v1.Deployment

	kubeClient client.Client
	cl kubernetes.Interface
)

var _ = BeforeSuite(func() {
	var err error
	var data []byte
	var c v1beta1.CustomResourceDefinition
	var d v1.Deployment

	By("Setup kube clients")
	kubeClient, cl, err = getClient()
	Expect(err).To(Succeed())
	cl.AppsV1()

	By("Read and Load CRD")
	data, err = ioutil.ReadFile( "../../deploy/crds/app.redislabs.com_proxyvalidatingtypes_crd.yaml")
	Expect(err).To(Succeed())
	Expect(yaml.NewYAMLToJSONDecoder(bytes.NewReader(data)).Decode(&c)).To(Succeed())
	Expect(kubeClient.Create(context.TODO(), &c)).To(Succeed())
	crd = &c

	By("Read and Load Operator")
	data, err = ioutil.ReadFile("../../deploy/operator.yaml")
	Expect(err).To(Succeed())
	Expect(yaml.NewYAMLToJSONDecoder(bytes.NewReader(data)).Decode(&d)).To(Succeed())
	d.Namespace = "default" // FIXME: make configurable
	Expect(kubeClient.Create(context.TODO(), &d)).To(Succeed())
	deploy = &d
	Expect(waitForDeployment(d)).To(Succeed())
})

var _ = AfterSuite(func() {
	// unload pod running operator
	if deploy != nil {
		Expect(kubeClient.Delete(context.TODO(), deploy)).To(Succeed())
	}

	// unload CRD
	if crd != nil {
		Expect(kubeClient.Delete(context.TODO(), crd)).To(Succeed())
	}
})