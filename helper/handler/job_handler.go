package handler

import (
	"autopilot-helper/helper/pkg/model"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

type JobHandler struct {
	DbManager *model.DbManager
}

func (h *JobHandler) Handle(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Receive request now")

	// Parse input data
	var inputData struct {
		SocketId string `json:"socketId"`
		Language string `json:"language"`
		Code     string `json:"code"`
		CodeId   string `json:"codeId"`
		Task     string `json:"task"`
		UserId   string `json:"userId"`
	}
	if err := json.NewDecoder(req.Body).Decode(&inputData); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Generate a random UUID for jobId
	jobId := uuid.New().String()

	err := h.DbManager.InsertJobData(inputData.SocketId, inputData.Language, inputData.Code, inputData.CodeId, inputData.Task, inputData.UserId, jobId, "in_queue")
	if err != nil {
		fmt.Printf("Error saving data to the database: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Respond with success message
	response := map[string]interface{}{
		"jobId": jobId,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
