package model

import (
	"time"
)

/*
{"code_name":"1","env":"dev","report_url":"https://example.com","started_at":"2023-06-12T11:53:37.038726Z","finished_at":"2023-06-12T11:53:37.038726Z","result":"pass","run_group_id":1,"test_result":"pass","real_result":"pass"}
*/
type RunResult struct {
	Id         int64     `db:"id" json:"id,omitempty"`
	CodeName   string    `db:"code_name" json:"code_name,omitempty"`
	Type       string    `db:"type" json:"type,omitempty"`
	Env        string    `db:"env" json:"env,omitempty"`
	ReportUrl  string    `db:"report_url" json:"report_url,omitempty"`
	StartedAt  time.Time `db:"started_at" json:"started_at,omitempty"`
	FinishedAt time.Time `db:"finished_at" json:"finished_at,omitempty"`
	Result     string    `db:"result" json:"result,omitempty"` // pass | fail
	RunGroupId int64     `db:"run_group_id" json:"run_group_id,omitempty"`
	TestResult string    `db:"test_result" json:"test_result,omitempty"`
	RealResult string    `db:"real_result" json:"real_result,omitempty"`
	BuildUrl   string    `db:"build_url" json:"build_url,omitempty"`
	TestId     string    `db:"test_id" json:"test_id,omitempty"` // id of test in the test file
}

/*
{"name":"schedule-fdsafds","env":"dev","report_url":"http://report-autopilot.shopbase.dev","started_at":"2023-06-12T13:31:54.151958Z","finish_at":"2023-06-12T13:31:54.151958Z","result":"pass"}
*/
type RunGroupResult struct {
	Id         int64     `db:"id" json:"id,omitempty"`
	Name       string    `db:"name" json:"name,omitempty"`
	Env        string    `db:"env" json:"env,omitempty"`
	ReportUrl  string    `db:"-" json:"report_url,omitempty"`
	StartedAt  time.Time `db:"started_at" json:"started_at,omitempty"`
	FinishedAt time.Time `db:"finish_at" json:"finished_at,omitempty"`
	Result     string    `db:"result" json:"result,omitempty"`
	JobId      int64     `db:"job_id" json:"job_id,omitempty"`
	TestResult string    `db:"test_result" json:"test_result,omitempty"`
}

type RunGroupJob struct {
	RunGroupID     *int    `json:"rungroup_id" db:"rungroup_id"`
	JobID          int     `json:"job_id" db:"job_id"`
	JobName        *string `json:"job_name" db:"job_name"`
	JobCallingCode *string `json:"job_calling_code" db:"job_calling_code"`
}

type RunGroupResultRequest struct {
	RunGroupResult
	Passed           int64  `db:"passed" json:"passed,omitempty"`
	Failed           int64  `db:"failed" json:"failed,omitempty"`
	TimedOut         int64  `db:"timeout" json:"timeout,omitempty"`
	Skipped          int64  `db:"skipped" json:"skipped,omitempty"`
	RunningUserEmail string `db:"running_user_email" json:"running_user_email"`
}

type EnvConfig struct {
	CodeName string `db:"code_name" json:"code_name,omitempty"`
	Env      string `db:"env" json:"env,omitempty"`
	Data     string `db:"data" json:"data,omitempty"`
}
