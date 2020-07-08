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
