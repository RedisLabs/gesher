package type_controller_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/avast/retry-go"
	"github.com/googleapis/gnostic/compiler"

	"github.com/redislabs/gesher/pkg/controller/proxyvalidatingtype"
	"k8s.io/api/admissionregistration/v1beta1"
	v1 "k8s.io/api/apps/v1"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/redislabs/gesher/pkg/apis"
	"github.com/redislabs/gesher/pkg/apis/app/v1alpha1"

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	. "github.com/onsi/ginkgo"
)

const (
	webhookResourceName = proxyvalidatingtype.ProxyWebhookName
)

func getClient() (client.Client, kubernetes.Interface, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, nil, err
	}

	mapper, err := apiutil.NewDynamicRESTMapper(cfg)
	if err != nil {
		return nil, nil, err
	}

	err = apis.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, nil, err
	}

	err = apiext.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, nil, err
	}

	objClient, err := client.New(cfg, client.Options{Scheme: scheme.Scheme, Mapper: mapper})
	if err != nil {
		return nil, nil, err
	}
	cl, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, nil, err
	}

	return objClient, cl, nil
}

func waitForDeployment(deploy v1.Deployment) error {
	return retry.Do(
		func() error {
			d := &v1.Deployment{}

			d, err := cl.AppsV1().Deployments(deploy.Namespace).Get(context.TODO(), deploy.Name, metav1.GetOptions{})
			// commented out as wasn't populating status, but would prefer to use if can be made to work
			//err := kubeClient.Get(context.TODO(), client.ObjectKey{Name: deploy.Name, Namespace: deploy.Namespace}, d)
			if err == nil {
				if d.Status.ReadyReplicas == 1 {
					return nil
				}
				fmt.Fprintf(GinkgoWriter, "%+v\n", d)
			} else if !apierrors.IsNotFound(err) {
				return retry.Unrecoverable(err)
			}

			fmt.Fprintf(GinkgoWriter, "failing d = %+v, err = %v\n", d, err)

			return fmt.Errorf("%v not available yet", deploy.Name)
		}, retry.Delay(5*time.Second),
	)
}

func verifyEmpty() error {
	return retry.Do(
		func() error {
			item := &v1beta1.ValidatingWebhookConfiguration{}
			err := kubeClient.Get(context.TODO(), client.ObjectKey{Name: webhookResourceName}, item)
			if err != nil {
				return retry.Unrecoverable(err)
			}

			switch len(item.Webhooks) {
			case 0:
				return nil
			case 1:
				if len(item.Webhooks[0].Rules) == 0 {
					return nil
				}
				return errors.New("expected no Rules")
			default: /**/
				return errors.New("more than one webhook specified")
			}
		}, retry.Delay(5*time.Second),
	)
}

func verifyApplied(pt *v1alpha1.ProxyValidatingType) error {
	prevGen := pt.Status.ObservedGeneration
	name := pt.Name

	return retry.Do(
		func() error {
			err := kubeClient.Get(context.TODO(), client.ObjectKey{Name: name}, pt)
			if err != nil {
				return retry.Unrecoverable(err)
			}
			if prevGen != pt.Status.ObservedGeneration && pt.Generation == pt.Status.ObservedGeneration {
				return nil
			}

			return errors.New("operator hasn't updated generation in status yet")
		},
		retry.Delay(5*time.Second),
	)
}

func verifyDeleted(pt *v1alpha1.ProxyValidatingType) error {
	name := pt.Name

	return retry.Do(
		func() error {
			err := kubeClient.Get(context.TODO(), client.ObjectKey{Name: name}, pt)
			if err == nil {
				return fmt.Errorf("%v not deleted yet", pt.Name)
			}
			if !apierrors.IsNotFound(err) {
				return retry.Unrecoverable(err)
			}

			return nil
		}, retry.Delay(5*time.Second),
	)
}

func validateInWebhook(ptList []*v1alpha1.ProxyValidatingType) error {
	item := &v1beta1.ValidatingWebhookConfiguration{}
	err := kubeClient.Get(context.TODO(), client.ObjectKey{Name: webhookResourceName}, item)
	if err != nil {
		return err
	}

	if len(item.Webhooks) != 1 {
		return errors.New("expected only a single webhook")
	}

	for _, pt := range ptList {
		if !proxyValidatingTypeExists(pt, item.Webhooks[0].Rules) {
			return fmt.Errorf("couldn't validate %+v in %+v", pt, item.Webhooks[0].Rules)
		}
	}

	return nil
}

func validateNotInWebhook(ptList []*v1alpha1.ProxyValidatingType) error {
	item := &v1beta1.ValidatingWebhookConfiguration{}
	err := kubeClient.Get(context.TODO(), client.ObjectKey{Name: webhookResourceName}, item)
	if err != nil {
		return err
	}

	if len(item.Webhooks) != 1 {
		return errors.New("expected only a single webhook")
	}

	for _, pt := range ptList {
		if proxyValidatingTypeExists(pt, item.Webhooks[0].Rules) {
			return fmt.Errorf("%+v still exists in %+v", pt, item.Webhooks[0].Rules)
		}
	}

	return nil
}

func proxyValidatingTypeExists(pt *v1alpha1.ProxyValidatingType, rules []v1beta1.RuleWithOperations) bool {
	for _, pType := range pt.Spec.Types {
		for _, group := range pType.APIGroups {
			for _, version := range pType.APIVersions {
				for _, resource := range pType.Resources {
				loop:
					for _, op := range pType.Operations {
						for _, rule := range rules {
							if compiler.StringArrayContainsValue(rule.APIGroups, group) &&
								compiler.StringArrayContainsValue(rule.APIVersions, version) &&
								compiler.StringArrayContainsValue(rule.Resources, resource) &&
								OpArrayContainsValues(rule.Operations, op) {
								// found a match, don't have to check anymore webhook rules, continue checking pType
								continue loop
							}
						}
						// only way to hit this should be if exhaust rules without finding a match
						return false
					}
				}
			}
		}
	}

	return true
}

func OpArrayContainsValues(operations []v1beta1.OperationType, op v1beta1.OperationType) bool {
	for _, operation := range operations {
		if op == operation {
			return true
		}
	}

	return false
}
