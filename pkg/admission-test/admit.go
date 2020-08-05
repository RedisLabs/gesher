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
	"k8s.io/klog"
	"net/http"

	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	AdmissionKey   = "admission-allow"
	AdmissionAllow = "true"
)

type Object struct {
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
}

func init() {
	http.HandleFunc("/admission", validate)
}

func validate(w http.ResponseWriter, r *http.Request) {
	Serve(w, r, admit)
}

func admit(review v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	klog.Infof("admit: enter")
	switch review.Request.Kind.Kind {
	default: // this handles everything
		var obj Object

		err := json.Unmarshal(review.Request.Object.Raw, &obj)
		if err != nil {
			return UnmarsallError(review.Request.Object.Raw, "admission_test.Object", err)
		}

		klog.Infof("obj = %+v\n", obj)

		val := obj.Labels[AdmissionKey]
		if val != AdmissionAllow {
			return GenericError(fmt.Errorf("%v label not set to allow this resource", AdmissionKey))
		}

		return Approved()
	}
}
