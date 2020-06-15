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

	log.Info(fmt.Sprintf("%v %v", requestedAdmissionReview, responseAdmissionReview))

	deserializer := apiserver.Codecs.UniversalDeserializer()

	if _, _, err := deserializer.Decode(body, nil, &requestedAdmissionReview); err != nil {
		log.Error(err, "deserializer failed")
		responseAdmissionReview.Response = errToAdmissionResponse(err)
	} else {
		webhooks := findWebhooks(requestedAdmissionReview.Request)
		responseAdmissionReview.Response = checkWebhooks(webhooks, r, bytes.NewReader(body))
	}

	// Return the same UID
	responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID

	log.Info(fmt.Sprintf("sending response: %v", responseAdmissionReview.Response))

	respBytes, err := json.Marshal(responseAdmissionReview)
	if err != nil {
		log.Error(err, "json marshall failed")
	}
	if _, err := w.Write(respBytes); err != nil {
		log.Error(err, "http response write failed")
	}
}
