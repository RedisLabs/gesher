package namespacedvalidatingproxy

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/redislabs/gesher/pkg/apis/app/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type observeState struct {
	customResource *v1alpha1.NamespacedValidatingProxy
}

func observe(kubeClient client.Client, request reconcile.Request, logger logr.Logger) (*observeState, error) {
	ret := &observeState{
		customResource: &v1alpha1.NamespacedValidatingProxy{},
	}

	err := kubeClient.Get(context.TODO(), request.NamespacedName, ret.customResource)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

