package handler

import (
	"autopilot-helper/helper/pkg/model"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type ChallengesHandler struct {
	DbManager *model.DbManager
}

type Video struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

type ChallengeResponse struct {
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
		Status string `json:"status"`
	} `json:"code"`
}

func (h *ChallengesHandler) Handle(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Receive request now")
	challenges, err := h.DbManager.GetAllChallenges()
	if err != nil {
		fmt.Printf("Cannot get challenges, err: %v",
			err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create a slice of ChallengeResponse to hold the transformed challenges
	var challengeResponses []ChallengeResponse

	// Loop through the challenges and transform them to ChallengeResponse
	for _, challenge := range challenges {
		var videos []Video
		// Parse the videos field from the challenge (assuming videos is a comma-separated string)
		err := json.Unmarshal([]byte(challenge.Videos), &videos)
		if err != nil {
			fmt.Println("Error parsing videos: ", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		challengeResp := ChallengeResponse{
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
				Status string `json:"status"`
			}{
				Status: "failed", // Set the status value based on your logic
			},
		}

		challengeResponses = append(challengeResponses, challengeResp)
	}

	respBytes, err := json.Marshal(challengeResponses)
	if err != nil {
		fmt.Println("Error marshaling response: ", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(respBytes)
	if err != nil {
		fmt.Println("Write byte error", err.Error())
		return
	}
}
