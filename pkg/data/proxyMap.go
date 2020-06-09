package data

import (
	"sync"

	"k8s.io/apimachinery/pkg/types"
	admissionv1beta1 "k8s.io/api/admissionregistration/v1beta1"
)

//Namespace->apiGroup->apiVersion->resource->operation->uid->webhook specy
type ProxyMap map[string]map[string]map[string]map[string]map[string]map[types.UID]admissionv1beta1.ValidatingWebhook

var (
	AdmissionProxyMap     = make(ProxyMap)
	AdmisisonProxyMapLock = sync.RWMutex{}
)
