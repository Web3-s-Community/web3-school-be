package handler

import (
	"autopilot-helper/helper/pkg/model"
	"encoding/json"
	"fmt"
	"net/http"
)

type ChallengeDetailHandler struct {
	DbManager *model.DbManager
}

func (h *ChallengeDetailHandler) Handle(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Receive request now")

	// Get the 'id' parameter from the URL
	challengeID := req.URL.Query().Get("id")

	challenge, err := h.DbManager.GetChallengeByID(challengeID)
	if err != nil {
		fmt.Printf("Cannot get challenge with ID %s, err: %v", challengeID, err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if challenge == nil {
		http.NotFound(w, req)
		return
	}

	respBytes, _ := json.Marshal(challenge)

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(respBytes)
	if err != nil {
		fmt.Println("Write byte error", err.Error())
		return
	}
}
