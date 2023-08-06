package model

import (
	"fmt"
	"testing"
)

func TestCallAlert(t *testing.T) {
	alert := NewCallAlert()
	alertInfo := AlertInfo{
		AlertCode: "sre-qe-smoke-test",
		CaseCode:  "SB_DEMO_05",
		JobName:   "Golang test job",
		ReportUrl: "https://google.com",
	}

	err := alert.MakeCall(&alertInfo)
	fmt.Println(err)

}
