package model

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestModel(t *testing.T) {
	m := RunResult{
		CodeName:   "a_11_37",
		Type:       "automation",
		Env:        "dev",
		ReportUrl:  "https://example.com",
		StartedAt:  time.Now().UTC(),
		FinishedAt: time.Now().UTC(),
		Result:     "pass",
		RunGroupId: 1,
		TestResult: "dsfdsfdsaf",
		RealResult: "pass",
		BuildUrl:   "https://autopilot.shopbase.dev",
	}

	mg := RunGroupResult{
		Name:       "schedule-fdsafds",
		Env:        "dev",
		ReportUrl:  "http://report-autopilot.shopbase.dev",
		StartedAt:  time.Now().UTC(),
		FinishedAt: time.Now().UTC(),
		Result:     "pass",
	}
	b1, _ := json.Marshal(m)
	b2, _ := json.Marshal(mg)
	fmt.Println(string(b1))
	fmt.Println(string(b2))
	t.Error()
}
