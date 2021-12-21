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

package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/operator-framework/operator-lib/leader"
	"github.com/redislabs/gesher/cmd/manager/flags"
	"github.com/redislabs/gesher/pkg/common"
	"github.com/redislabs/gesher/pkg/tls_manager"
	"k8s.io/client-go/kubernetes"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	zapcr "sigs.k8s.io/controller-runtime/pkg/log/zap"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"

	admission_proxy "github.com/redislabs/gesher/pkg/admission-proxy"
	"github.com/redislabs/gesher/pkg/apis"
	"github.com/redislabs/gesher/pkg/controller"
	"github.com/redislabs/gesher/version"

	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

// Change below variables to serve metrics on different host or port.
var (
	metricsHost       = "0.0.0.0"
	metricsPort int32 = 8383
)
var log = logf.Log.WithName("cmd")

func printVersion() {
	log.Info(fmt.Sprintf("Operator Version: %s", version.Version))
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
}

func getWatchNamespace() (string, error) {
	const watchNamespaceEnvVar = "WATCH_NAMESPACE"
	ns, found := os.LookupEnv(watchNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", watchNamespaceEnvVar)
	}
	return ns, nil
}

func main() {
	// Zap logger setup.
	configLog := zap.NewProductionEncoderConfig()
	configLog.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(configLog)

	opts := zapcr.Options{}
	opts.BindFlags(flag.CommandLine)

	flag.Parse()
	logger := zapcr.New(zapcr.UseFlagOptions(&opts), zapcr.Encoder(encoder))
	logf.SetLogger(logger)

	printVersion()

	namespace, err := getWatchNamespace()
	if err != nil {
		log.Error(err, "Failed to get watch namespace")
		os.Exit(1)
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Setup TLS
	err = setupTLS(cfg)
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	ctx := context.TODO()
	// Become the leader before proceeding
	err = leader.Become(ctx, "gesher-lock")
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Set default manager options
	options := manager.Options{
		Namespace:          namespace,
		MetricsBindAddress: fmt.Sprintf("%s:%d", metricsHost, metricsPort),
	}

	// Add support for MultiNamespace set in WATCH_NAMESPACE (e.g ns1,ns2)
	// Note that this is not intended to be used for excluding namespaces, this is better done via a Predicate
	// Also note that you may face performance issues when using this with a high number of namespaces.
	// More Info: https://godoc.org/github.com/kubernetes-sigs/controller-runtime/pkg/cache#MultiNamespacedCacheBuilder
	if strings.Contains(namespace, ",") {
		options.Namespace = ""
		options.NewCache = cache.MultiNamespacedCacheBuilder(strings.Split(namespace, ","))
	}

	// Create a new manager to provide shared dependencies and start components
	mgr, err := manager.New(cfg, options)
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	log.Info("Registering Components.")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	setupWebhook(mgr)

	log.Info("Starting the Cmd.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Manager exited non-zero")
		os.Exit(1)
	}
}

// Simple health/liveness endpoint
type Healthz struct{}

func (h Healthz) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("ok"))
}

func setupWebhook(mgr manager.Manager) {
	// TODO: hack to not annoy linter to enable code to remain
	//	enableWebhook := os.Getenv("ENABLE_WEBHOOK")
	//	if enableWebhook == "yes" {
	server := mgr.GetWebhookServer()

	server.CertDir = common.CertDir
	server.CertName = common.CertPem
	server.KeyName = common.PrivPem
	server.Port = 8443

	// register objects that serve the 2 primary endpoints
	server.Register("/healthz", &Healthz{})
	server.Register(common.ProxyPath, &admission_proxy.Handler{})
	//	}
}

func setupTLS(cfg *rest.Config) error {
	client := kubernetes.NewForConfigOrDie(cfg)

	priv, cert, err := tls_manager.GenerateTLS(client, *flags.Namespace, *flags.Service, *flags.TlsSecret)
	if err != nil {
		return err
	}

	err = os.MkdirAll(common.CertDir, 0700)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(common.CertDir, common.PrivPem), priv, 0600)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(common.CertDir, common.CertPem), cert, 0600)
	if err != nil {
		return err
	}

	return nil
}
