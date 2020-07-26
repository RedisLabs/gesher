package admission_proxy

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"k8s.io/api/admission/v1beta1"
	admv1beta1 "k8s.io/api/admissionregistration/v1beta1"

	"github.com/redislabs/gesher/pkg/controller/namespacedvalidatingproxy"

)

func findWebhooks(request *v1beta1.AdmissionRequest) []namespacedvalidatingproxy.WebhookConfig {
	op := admv1beta1.OperationType(request.Operation)

	return namespacedvalidatingproxy.EndpointData.Get(request.Namespace, request.RequestKind, op)
}

// code is inspired by k8s.io/apiserver/pkg/admission/plugin/webhook/validating/dispatcher.go
func checkWebhooks(webhooks []namespacedvalidatingproxy.WebhookConfig, r *http.Request, body *bytes.Reader) *v1beta1.AdmissionResponse {
	if len(webhooks) == 0 {
		return approved()
	}

	wg := &sync.WaitGroup{}
	errCh := make(chan error, len(webhooks))
	wg.Add(len(webhooks))
	for _, webhook := range webhooks {
		go doWebhook(webhook, wg, r, body, errCh)
	}

	wg.Wait()
	close(errCh)

	var errs []error
	for e := range errCh {
		if e != nil {
			errs = append(errs, e)
		}
	}
	if len(errs) == 0 {
		return approved()
	}

	if len(errs) > 1 {
		// TODO: merge status errors; until then, just return the first one.
		log.V(3).Info("TODO: merge status errors; until then, just return the first one.")
	}

	return errToAdmissionResponse(errs[0])
}

func doWebhook(webhook namespacedvalidatingproxy.WebhookConfig, wg *sync.WaitGroup, r *http.Request, body *bytes.Reader, errCh chan error) {
	defer wg.Done()

	service := webhook.ClientConfig.Service

	url := fmt.Sprintf("https://%v.%v:%v%v", service.Name, service.Namespace, service.Port, service.Path)

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(webhook.ClientConfig.CABundle)

	client := &http.Client{
		Timeout: time.Duration(webhook.TimeoutSecs) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}

	req, _ := http.NewRequestWithContext(context.TODO(), "POST", url, body)
	for k, v := range r.Header {
		for _, s := range v {
			req.Header.Add(k, s)
		}
	}

	resp, err := client.Do(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	failurePolicy := admv1beta1.Fail
	err = toFailure("webhook", resp, err, failurePolicy)

	errCh <- err
}

func toFailure(name string, resp *http.Response, httpErr error, failurePolicy admv1beta1.FailurePolicyType) error {
	log.V(2).Info(fmt.Sprintf("toFailure: %v: httpErr = %v", name, httpErr))
	if httpErr != nil {
		return errToFailure(name, fmt.Errorf("http error: %v", httpErr), failurePolicy)
	}

	if resp.Body == nil {
		return errToFailure(name, errors.New("response body is nil"), failurePolicy)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errToFailure(name, fmt.Errorf("ReadAll failed: %v", err), failurePolicy)
	}

	log.V(2).Info(fmt.Sprintf("toFailure: resp.Body = %v", string(data)))

	var responseAdmissionReview v1beta1.AdmissionReview
	err = json.Unmarshal(data, &responseAdmissionReview)
	if err != nil {
		return errToFailure(name, fmt.Errorf("json unmarshall failed: %v", err), failurePolicy)
	}

	log.V(2).Info(fmt.Sprintf("toFailure: unmarshalled response = %+v\n", responseAdmissionReview))

	if !responseAdmissionReview.Response.Allowed {
		return fmt.Errorf("proxied webhook %v denied the request: %v", name, responseAdmissionReview.Response.Result.Message)
	}

	log.V(2).Info("toFailure: passed all test")

	return nil
}

func errToFailure(name string, err error, failurePolicy admv1beta1.FailurePolicyType) error {
	switch failurePolicy {
	case admv1beta1.Fail:
		log.V(1).Info(fmt.Sprintf("err = %v and FailurePolicy == Fail", err))
		return fmt.Errorf("proxied webhook %v failed: %v", name, err)
	default:
		log.V(1).Info(fmt.Sprintf("err = %v and FailurePolicy == Ignore", err))
		return nil
	}
}
