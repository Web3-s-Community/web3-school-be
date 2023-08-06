package handler

import (
	"net/http"
)

type DemoHandler struct {
}

func (h *DemoHandler) Handle(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("content-type", "application/json")
	w.Write([]byte(`{"hello": "world"}`))
}
