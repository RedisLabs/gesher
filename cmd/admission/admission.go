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
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/redislabs/gesher/pkg/tls_manager"

	_ "github.com/redislabs/gesher/pkg/admission-test"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var (
	port = flag.Int("port", 9443, "port to run https server on")
)

func init() {
	http.HandleFunc("/healthz", healthz)
}

func main() {
	defer klog.Flush()

	flag.Parse()

	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	cl := kubernetes.NewForConfigOrDie(cfg)

	namespace := os.Getenv("POD_NAMESPACE")

	privKey, cert, err := tls_manager.GenerateTLS(cl, namespace, "admission-test", "admission-test")
	if err != nil {
		klog.Infof("GenerateTLS failed")
		os.Exit(1)
	}

	klog.Infof("Setting Up Web Server")

	server := &http.Server{
		Addr:      fmt.Sprintf(":%v", *port),
		TLSConfig: configTLS(privKey, cert),
	}

	err = server.ListenAndServeTLS("", "")
	if err != nil {
		klog.Errorf("ListenAndServeTLS failed: %v", err)
		os.Exit(1)
	}

	// Shouldn't reach here on normal execution
}

func healthz(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("ok"))
}

func configTLS(privateKey, cert []byte) *tls.Config {
	sCert, err := tls.X509KeyPair(cert, privateKey)
	if err != nil {
		klog.Fatal(err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{sCert},
		// This is from k8s example code and leaving here if necessary for future
		// TODO: uses mutual tls after we agree on what cert the apiserver should use.
		// ClientAuth:   tls.RequireAndVerifyClientCert,
	}
}
