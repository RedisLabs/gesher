package admission_proxy

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"k8s.io/api/admission/v1beta1"
	admv1beta1 "k8s.io/api/admissionregistration/v1beta1"
)

// temporary dumb webhook handler
type Webhook map[string]string

const (
	serviceName = "service-name"
	namespace   = "namespace"
	port        = "port"
	path        = "path"
	caBundle    = "ca-bundle"
)

func findWebhooks(request *v1beta1.AdmissionRequest) []Webhook {
	webhook := make(Webhook)

	if webhook[serviceName] = os.Getenv("ADM_SERVICE_NAME"); webhook[serviceName] == "" {
		panic("failed to read env variable ADM_SERVICE_NAME")
	}
	if webhook[namespace] = os.Getenv("ADM_SERVICE_NAMESPACE"); webhook[namespace] == "" {
		panic("failed to read env variable ADM_SERVICE_NAMESPACE")
	}
	if webhook[path] = os.Getenv("ADM_SERVICE_ENDPOINT"); webhook[path] == "" {
		panic("failed to read env variable ADM_SERVICE_ENDPOINT")
	}
	encoded := os.Getenv("ADM_SERVICE_CABUNDLE")
	if encoded == "" {
		panic("failed to read env variable ADM_SERVICE_CABUNDLE")
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		panic("failed to base64 decode cabundle")
	}
	webhook[caBundle] = string(data)

	if webhook[port] = os.Getenv("ADM_SERVICE_PORT"); webhook[port] == "" {
		panic("failed to read env variable ADM_SERVICE_CABUNDLE")
	}
	if _, err := strconv.Atoi(webhook[port]); err != nil {
		panic(fmt.Errorf("failed to convert port to an int: %v", err))
	}

	return []Webhook{webhook}
}

func checkWebhooks(webhooks []Webhook, r *http.Request, body *bytes.Reader) *v1beta1.AdmissionResponse {
	if len(webhooks) == 0 {
		return approved()
	}

	webhook := webhooks[0]

	url := fmt.Sprintf("https://%v.%v:%v%v", webhook[serviceName], webhook[namespace], webhook[port], webhook[path])

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM([]byte(webhook[caBundle]))

	client := &http.Client{
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
	if err != nil {
		defer resp.Body.Close()
	}
	failurePolicy := admv1beta1.Fail
	err = toFailure(resp, err, failurePolicy)

	if err != nil {
		return errToAdmissionResponse(err)
	}

	return approved()
}

func toFailure(resp *http.Response, httpErr error, failurePolicy admv1beta1.FailurePolicyType) error {
	log.Info(fmt.Sprintf("toFailure: httpErr = %v", httpErr))
	if httpErr != nil {
		if failurePolicy == admv1beta1.Fail {
			log.Info("returning httpErr as failurePolicy = Fail")
			return httpErr
		} else {
			log.Info("httpErr, but returning nil as failurePolicy = Ignore")
			return nil
		}
	}

	if resp.Body == nil {
		if failurePolicy == admv1beta1.Fail {
			log.Info("body is nil, returning error as failurePolicy = Fail")
			return errors.New("empty response")
		} else {
			log.Info("body is nil, returning nil as failurePolicy = Ignore")
			return nil
		}
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		if failurePolicy == admv1beta1.Fail {
			log.Info("failed to read body, returning error as failurePolicy = Fail")
			return err
		} else {
			log.Info("failed to read body, returning nil as failurePolicy = Ignore")
			return nil
		}
	}

	log.Info(fmt.Sprintf("toFailure: resp.Body = %v", string(data)))

	var responseAdmissionReview v1beta1.AdmissionReview
	err = json.Unmarshal(data, &responseAdmissionReview)
	if err != nil {
		if failurePolicy == admv1beta1.Fail {
			log.Info("failed to json unmarshall, returning error as failurePolicy = Fail")
			return err
		} else {
			log.Info("failed to json unmarshall, returning nil as failurePolicy = Ignore")
			return nil
		}
	}

	log.Info(fmt.Sprintf("toFailure: unmarshalled response = %+v\n", responseAdmissionReview))

	if !responseAdmissionReview.Response.Allowed {
		return errors.New(responseAdmissionReview.Response.Result.Message)
	}

	log.Info("toFailure: passed all test")

	return nil
}