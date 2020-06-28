package tls_manager

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"

	"github.com/pkg/errors"

	v1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

type TLSManager interface {
	HasKey() bool
	CreateKey() error
	GetKey() ([]byte, []byte, error)
	ConfigTLS([]byte, []byte) *tls.Config
}

const (
	privateKeySecretKey = "privateKey"
	certSecretKey       = "cert"
)

type kubeTLSManager struct {
	kubeClient kubernetes.Interface
	namespace  string
	name       string
	ips        []net.IP
	dnsNames   []string
}

//NewTLSManager
func NewTLSManager(kubeClient kubernetes.Interface, namespace, name string, ips []net.IP, dns []string) TLSManager {
	return &kubeTLSManager{
		kubeClient: kubeClient,
		namespace:  namespace,
		name:       name,
		ips:        ips,
		dnsNames:   dns,
	}
}

func (k *kubeTLSManager) HasKey() bool {
	secret, err := k.getSecret()
	if err != nil {
		klog.Error(err, "hasKey's call to GetSecret failed")
		return false
	}

	return secret != nil
}

func (k *kubeTLSManager) CreateKey() error {
	if k.ips == nil && k.dnsNames == nil {
		return errors.New("can't create a keypair if ips and dnsNames are both nil")
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return errors.Wrap(err, "failed to generate key")
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(24 * 365 * 5 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		klog.Errorf("Failed to generate serial number: %v", err)
		return err
	}

	template := x509.Certificate{
		SerialNumber:       serialNumber,
		SignatureAlgorithm: x509.SHA256WithRSA,
		Subject: pkix.Name{
			CommonName:   k.name,
			Organization: []string{"RedisLabs Admission Control"},
		},
		NotBefore:   notBefore,
		NotAfter:    notAfter,
		IPAddresses: k.ips,
		DNSNames:    k.dnsNames,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, privateKey.Public(), privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %v", err)
	}

	cert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	pemdata := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: k.name,
		},
		Data: map[string][]byte{
			privateKeySecretKey: pemdata,
			certSecretKey:       cert,
		},
	}

	klog.Infof("CreateKey: namespace = %v, secret = %v", k.namespace, k.name)
	_, err = k.kubeClient.CoreV1().Secrets(k.namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		err = errors.Wrap(err, "createKey failed to create secret")
	}

	return err
}

func (k *kubeTLSManager) GetKey() ([]byte, []byte, error) {
	secret, err := k.getSecret()
	if err != nil {
		return nil, nil, errors.Wrap(err, "GetSecret failed")
	}
	if secret == nil {
		return nil, nil, errors.New("private key secret doesn't exit")
	}

	privKey, ok := secret.Data[privateKeySecretKey]
	if !ok {
		return nil, nil, errors.New("secret doesn't contain private key")
	}

	csr, ok := secret.Data[certSecretKey]
	if !ok {
		return nil, nil, errors.Wrap(err, "failed to get certificate")
	}

	return privKey, csr, nil
}

func (k *kubeTLSManager) getSecret() (*v1.Secret, error) {
	klog.Infof("getSecret: namespace = %v, secret = %v", k.namespace, k.name)

	secret, err := k.kubeClient.CoreV1().Secrets(k.namespace).Get(context.TODO(), k.name, metav1.GetOptions{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return nil, err
		} else {
			return nil, nil
		}
	}

	return secret, nil
}

func (k *kubeTLSManager) ConfigTLS(privateKey, cert []byte) *tls.Config {
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
