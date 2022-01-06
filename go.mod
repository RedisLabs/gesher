module github.com/redislabs/gesher

go 1.16

require (
	github.com/go-logr/logr v0.4.0
	github.com/googleapis/gnostic v0.5.5
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.15.0
	github.com/operator-framework/operator-lib v0.9.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.19.0
	k8s.io/api v0.22.2
	k8s.io/apiextensions-apiserver v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.10.0
)

replace github.com/operator-framework/api => github.com/operator-framework/api v0.11.0

replace sigs.k8s.io/controller-tools => sigs.k8s.io/controller-tools v0.7.0

replace github.com/spf13/cobra => github.com/spf13/cobra v1.2.0

replace sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.10.3
