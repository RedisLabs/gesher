module github.com/redislabs/gesher

go 1.15

require (
	github.com/go-logr/logr v0.4.0
	github.com/googleapis/gnostic v0.5.1
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/operator-framework/operator-lib v0.3.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.15.0
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	k8s.io/api v0.19.4
	k8s.io/apiextensions-apiserver v0.19.3
	k8s.io/apimachinery v0.19.4
	k8s.io/client-go v0.19.4
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.7.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.1
	google.golang.org/grpc => google.golang.org/grpc v1.26.0 // Requred by apiextensions-apiserve's version of etcd
	k8s.io/client-go => k8s.io/client-go v0.19.4 // Required by prometheus-operator
)
