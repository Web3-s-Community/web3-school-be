package model

type UpdateJobRequest struct {
	Id      int64  `json:"id" db:"id"`
	Crontab string `json:"crontab" db:"crontab"`
}

type InsertJobRequest struct {
	Name             string `json:"name" db:"name"`
	Crontab          string `json:"crontab" db:"crontab"`
	Enabled          bool   `json:"enabled" db:"enabled"`
	Mode             string `json:"mode" db:"mode"`
	Env              string `json:"env" db:"env"`
	CreatedById      int64  `json:"created_by_id" db:"created_by_id"`
	CreatedBySlackId string `json:"created_by_slack_id" db:"created_by_slack_id"`
	CreatedByEmail   string `json:"created_by_email" db:"created_by_email"`
	UpdatedAt        string `json:"updated_at" db:"updated_at"`
	Branch           string `json:"branch" db:"branch"`
}

type RemoveJobRequest struct {
	Id int64 `json:"id" db:"id"`
}
