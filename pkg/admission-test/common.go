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

package admission_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/apiserver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"net/http"
)

// admitFunc is the type we use for all of our validators and mutators
type admitFunc func(v1beta1.AdmissionReview) *v1beta1.AdmissionResponse

// Serve handles the http portion of a request prior to handing to an admit
// function
func Serve(w http.ResponseWriter, r *http.Request, admit admitFunc) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("contentType=%s, expect application/json", contentType)
		return
	}

	klog.V(3).Info(fmt.Sprintf("handling request: %s", body))

	// The AdmissionReview that was sent to the webhook
	requestedAdmissionReview := v1beta1.AdmissionReview{}

	// The AdmissionReview that will be returned
	responseAdmissionReview := v1beta1.AdmissionReview{}

	deserializer := apiserver.Codecs.UniversalDeserializer()
	name := "DecoderError"
	if _, _, err := deserializer.Decode(body, nil, &requestedAdmissionReview); err != nil {
		klog.Error(err)
		responseAdmissionReview.Response = toAdmissionResponse(err)
	} else {
		klog.V(1).Infof("Validating %v", requestedAdmissionReview.Request.Name)
		// pass to admitFunc
		responseAdmissionReview.Response = admit(requestedAdmissionReview)
	}

	// Return the same UID
	responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID

	klog.V(2).Info(fmt.Sprintf("sending response: %v", responseAdmissionReview.Response))
	if responseAdmissionReview.Response.Allowed {
		klog.V(1).Infof("%v: Approved", name)
	} else {
		klog.V(1).Infof("%v: Denied", name)
	}

	respBytes, err := json.Marshal(responseAdmissionReview)
	if err != nil {
		klog.Error(err)
	}
	if _, err := w.Write(respBytes); err != nil {
		klog.Error(err)
	}
}

// toAdmissionResponse is a helper function to create an AdmissionResponse with an embedded error
func toAdmissionResponse(err error) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}

func UnknownKindError(kind metav1.GroupVersionKind) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Allowed: false,
		Result: &metav1.Status{
			Status:  "Failure",
			Message: fmt.Sprintf("This admission controller should not be getting a %+v", kind),
		},
	}
}

func UnmarsallError(raw []byte, t string, err error) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Allowed: false,
		Result: &metav1.Status{
			Status:  "Failure",
			Message: fmt.Sprintf("Couldn't unmarshall %v to type %v: %v", string(raw), t, err),
		},
	}
}

func GenericError(err error) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Allowed: false,
		Result: &metav1.Status{
			Status:  "Failure",
			Message: err.Error(),
		},
	}
}

func Approved() *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Allowed: true,
	}
}
