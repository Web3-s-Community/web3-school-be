package model

type InsertCaseJobRequest struct {
	CaseId int64 `json:"case_id" db:"case_id"`
	JobId  int64 `json:"job_id" db:"job_id"`
}

type RemoveCaseJobRequest struct {
	CaseId int64 `json:"case_id" db:"case_id"`
	JobId  int64 `json:"job_id" db:"job_id"`
}
