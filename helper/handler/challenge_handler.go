package handler

import (
	"autopilot-helper/helper/pkg/model"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type ChallengeDetailHandler struct {
	DbManager *model.DbManager
}

type ChallengeDetailResponse struct {
	ID         string   `json:"id"`
	Slug       string   `json:"slug"`
	Language   string   `json:"language"`
	Title      string   `json:"title"`
	Difficulty string   `json:"difficulty"`
	Free       bool     `json:"free"`
	Videos     []Video  `json:"videos"`
	Tags       []string `json:"tags"`
	Sort       int      `json:"sort"`
	Code       struct {
		Code   string `json:"code"`
		Status string `json:"status"`
	} `json:"code"`
	RawCode  string   `json:"raw_code"`
	Points   int64    `json:"points"`
	Prompt   string   `json:"prompt"`
	Starter  string   `json:"starter"`
	Hints    []string `json:"hints"`
	Tasks    []string `json:"tasks"`
	Solution string   `json:"solution"`
	Test     string   `json:"test"`
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

	var videos []Video
	// Parse the videos field from the challenge (assuming videos is a comma-separated string)
	err = json.Unmarshal([]byte(challenge.Videos), &videos)
	if err != nil {
		fmt.Println("Error parsing videos: ", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hints := strings.Split(challenge.Hints, ",")
	tasks := strings.Split(challenge.Tasks, ",")

	fmt.Println(challenge.Code)

	challengeResp := ChallengeDetailResponse{
		ID:         challenge.ID,
		Slug:       challenge.Slug,
		Language:   challenge.Language,
		Title:      challenge.Title,
		Difficulty: challenge.Difficulty,
		Free:       challenge.Free,
		Videos:     videos,
		Tags:       strings.Split(challenge.Tags, ","),
		Sort:       1, // You may set the sort value according to your requirement
		Code: struct {
			Code   string `json:"code"`
			Status string `json:"status"`
		}{
			Status: "failed",
			Code:   challenge.Code,
		},
		RawCode:  challenge.Code,
		Points:   challenge.Points,
		Prompt:   challenge.Prompt,
		Starter:  challenge.Starter,
		Hints:    hints,
		Tasks:    tasks,
		Solution: challenge.Solution,
		Test:     challenge.Test,
	}

	respBytes, _ := json.Marshal(challengeResp)

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(respBytes)
	if err != nil {
		fmt.Println("Write byte error", err.Error())
		return
	}
}
