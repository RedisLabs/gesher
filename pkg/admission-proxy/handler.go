package admission_proxy

import "net/http"

type Handler struct {}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(501)
	_, _ = w.Write([]byte("Implement Me"))
}

