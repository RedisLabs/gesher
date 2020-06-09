package data

import (
	"sync"

	"k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/types"
)

//apiGroup->apiVersion->resource->operation->uid
type TypeMap map[string]map[string]map[string]map[v1beta1.OperationType]map[types.UID]bool

var (
	ProxyTypes    = make(TypeMap)
	ProxyTypeLock sync.RWMutex
)
