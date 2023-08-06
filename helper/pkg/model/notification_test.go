package model

import (
	"fmt"
	"testing"
	"time"
)

func TestNotify2(t *testing.T) {
	notificator := NewNotificator()
	noti := Notification{
		Passed:      1,
		Failed:      1,
		TimedOut:    1,
		Skipped:     1,
		SlackUserID: "U0NCL405Q",
		JobName:     "fdsaf",
		RunMode:     "multiple-time",
		StartedTime: time.Now().Add(time.Hour * -1),
		ThJobID:     280,
		Env:         "prod",
		BuildURL:    "https://google.com",
	}
	fmt.Println(notificator.Notify(&noti))
	t.Error()
}
