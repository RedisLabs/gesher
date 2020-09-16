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

package admission_proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"k8s.io/api/admission/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/apiserver"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("handler")

type Handler struct{}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		log.Info(fmt.Sprintf("contentType=%s, expect application/json", contentType))
		return
	}

	// The AdmissionReview that was sent to the webhook
	requestedAdmissionReview := v1beta1.AdmissionReview{}

	// The AdmissionReview that will be returned
	responseAdmissionReview := v1beta1.AdmissionReview{}

	deserializer := apiserver.Codecs.UniversalDeserializer()

	if _, _, err := deserializer.Decode(body, nil, &requestedAdmissionReview); err != nil {
		log.Error(err, "deserializer failed")
		responseAdmissionReview.Response = errToAdmissionResponse(err)
	} else {
		log.V(2).Info(fmt.Sprintf("request = %+v", requestedAdmissionReview))
		webhooks := findWebhooks(requestedAdmissionReview.Request)
		log.V(2).Info(fmt.Sprintf("webhooks = %+v", webhooks))
		responseAdmissionReview.Response = checkWebhooks(webhooks, r, bytes.NewReader(body))
		log.V(2).Info(fmt.Sprintf("response = %+v", responseAdmissionReview.Response))
	}

	// Return the same UID
	responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID

	log.V(2).Info(fmt.Sprintf("sending response: %v", responseAdmissionReview.Response))

	respBytes, err := json.Marshal(responseAdmissionReview)
	if err != nil {
		log.Error(err, "json marshall failed")
	}
	if _, err := w.Write(respBytes); err != nil {
		log.Error(err, "http response write failed")
	}
}
