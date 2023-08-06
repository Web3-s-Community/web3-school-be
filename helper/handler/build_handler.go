package handler

import (
	"fmt"
	"net/http"
)

type DemoHandler struct {
}

func (h *DemoHandler) Handle(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Receive request now")
	w.Header().Add("content-type", "application/json")
	w.Write([]byte(`{"hello": "world"}`))
}
