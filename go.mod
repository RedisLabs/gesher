module github.com/redislabs/gesher

go 1.13

require (
	github.com/go-logr/logr v0.1.0
	github.com/operator-framework/operator-sdk v0.18.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.5.1
	github.com/avast/retry-go v2.6.0+incompatible
	github.com/googleapis/gnostic v0.3.1
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.18.2
	k8s.io/apiextensions-apiserver v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	google.golang.org/grpc => google.golang.org/grpc v1.26.0 // Requred by apiextensions-apiserve's version of etcd
	k8s.io/client-go => k8s.io/client-go v0.18.2 // Required by prometheus-operator
)
