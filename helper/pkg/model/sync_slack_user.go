package model

type SyncSlackUserInput struct {
	Email string `json:"email" db:"email"`
}

type SlackUser struct {
	SlackID  *string `json:"slack_id" db:"slack_id"`
	Email    string  `json:"email" db:"email"`
	IsActive int     `json:"is_active" db:"is_active"`
}
