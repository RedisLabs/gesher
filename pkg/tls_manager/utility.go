package tls_manager

import (
	"fmt"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"net"
	"os"
)

func GenerateTLS(client kubernetes.Interface, namespace, serviceName, secretName string) ([]byte, []byte, error) {
	var ips []net.IP

	serviceIp, err := net.LookupIP(fmt.Sprintf("%v.%v", serviceName, namespace))
	if err == nil {
		klog.V(1).Infof("serviceIp = %+v", serviceIp)
		ips = append(ips, serviceIp...)
	} else {
		klog.Warningf("couldn't resolve service ip: %v", err)
	}

	klog.Infof("Service IP = %v\n", serviceIp[0].String())

	ips, dnsNames := GetIPsAndNames(ips, serviceName, namespace)

	klog.Infof("dnsNames = %+v", dnsNames)

	tlsManager := NewTLSManager(client, namespace, secretName, ips, dnsNames)

	if !tlsManager.HasKey() {
		err := tlsManager.CreateKey()
		if err != nil {
			err = errors.Wrap(err, "failed to create key")
			return nil, nil, fmt.Errorf("err = %v", err)
		}
	}

	return tlsManager.GetKey()
}

func GetIPsAndNames(ips []net.IP, serviceName string, namespace string) ([]net.IP, []string) {
	podIp := os.Getenv("POD_IP")
	if podIp != "" {
		ip := net.ParseIP(podIp)
		if ip != nil {
			ips = append(ips, ip)
		} else {
			klog.Warningf("couldn't parse ip %v", podIp)
		}
	}

	dnsNames := []string{
		serviceName,
		fmt.Sprintf("%v.%v.svc.cluster.local", serviceName, namespace),
		fmt.Sprintf("%v.%v.svc", serviceName, namespace),
		fmt.Sprintf("%v.%v", serviceName, namespace),
	}

	return ips, dnsNames
}
