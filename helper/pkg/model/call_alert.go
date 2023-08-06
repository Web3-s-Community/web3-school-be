package model

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"time"
)

type CallAlert struct {
	BaseUrl string
}

func NewCallAlert() *CallAlert {

	return &CallAlert{
		BaseUrl: "https://metrics-hub.beeketing.net/product-hub/alert/notify",
	}
}

type AlertInfo struct {
	AlertCode string
	CaseCode  string
	CaseId    int64
	ReportUrl string
	JobName   string
}

func (c *CallAlert) MakeCall(info *AlertInfo) error {
	if info.AlertCode == "" {
		fmt.Println("Missing alert code")
		return errors.New("Missing alert code")
	}

	callUrl := fmt.Sprintf("%v/%v", c.BaseUrl, info.AlertCode)
	callBody := map[string]interface{}{
		"body": fmt.Sprintf("%v: report_url: %v, test_case_url: https://test-hub.ocg.to/admin/th/case/%v/change, job_name: %v",
			info.CaseCode,
			info.ReportUrl,
			info.CaseId,
			info.JobName),
	}

	b, _ := json.Marshal(callBody)

	callReq, err := http.NewRequest(http.MethodPost, callUrl, bytes.NewBuffer(b))
	if err != nil {
		fmt.Println("Cannot create new request: ", err)
		return err
	}

	b, _ = httputil.DumpRequest(callReq, true)
	fmt.Println("Created request for: ", callUrl, ": ", string(b))

	resp, err := NewClient().Do(callReq)
	if err != nil {
		fmt.Println("Cannot do request", err)
		return err
	}
	b, _ = httputil.DumpResponse(resp, true)
	fmt.Println("Dump request call alert: ", string(b))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		fmt.Println("Wrong return code from metrics-hub: ", resp.StatusCode)
		return errors.New(fmt.Sprintf("Wrong code from metrics-hub: %v", resp.StatusCode))
	}

	return nil
}

func NewClient() *http.Client {
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	var netClient = &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}
	return netClient
}
