package model

type UpdateCaseDataRequest struct {
	Id       int64  `json:"id" db:"id"`
	CaseId   int64  `json:"case_id" db:"case_id"`
	CaseCode string `json:"case_code" db:"case_code"`
	Data     string `json:"data" db:"data"`
	Env      string `json:"env" db:"env"`
	Branch   string `json:"branch" db:"branch"`
}

type InsertCaseDataRequest struct {
	CaseId   int64  `json:"case_id" db:"case_id"`
	CaseCode string `json:"case_code" db:"case_code"`
	Data     string `json:"data" db:"data"`
	Env      string `json:"env" db:"env"`
	Branch   string `json:"branch" db:"branch"`
}

type RemoveCaseDataRequest struct {
	CaseId int64  `json:"case_id" db:"case_id"`
	Env    string `json:"env" db:"env"`
	Branch string `json:"branch" db:"branch"`
}
