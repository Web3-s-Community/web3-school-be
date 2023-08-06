package model

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

type Notificator struct {
	slackClient *slack.Client
}

func NewNotificator() *Notificator {
	token := os.Getenv("SLACK_AUTH_TOKEN")
	appToken := os.Getenv("SLACK_APP_TOKEN")
	client := slack.New(token, slack.OptionDebug(true), slack.OptionAppLevelToken(appToken))
	return &Notificator{
		slackClient: client,
	}
}

func (n *Notificator) GetSlackClient() *slack.Client {
	return n.slackClient
}

type Notification struct {
	Passed       int64     `json:"passed,omitempty"`
	Failed       int64     `json:"failed,omitempty"`
	TimedOut     int64     `json:"timed_out,omitempty"`
	Skipped      int64     `json:"skipped,omitempty"`
	SlackUserID  string    `json:"slack_user_id,omitempty"`
	SlackLeadID  string    `json:"slack_lead_id,omitempty"`
	SlackWebhook string    `json:"slack_webhook,omitempty"`
	SlackChannel string    `json:"slack_channel,omitempty"`
	JobName      string    `json:"job_name,omitempty"`
	RunMode      string    `json:"run_mode,omitempty"`
	StartedTime  time.Time `json:"started_time,omitempty"`
	ThJobID      int64     `json:"job_id,omitempty"`
	Env          string    `json:"env,omitempty"`
	BuildURL     string    `json:"build_url,omitempty"`
	ThRunGroupId int64     `json:"run_group_id,omitempty"`
}

var DefaultSlackChannel = "qcev-run-result"

func (n *Notificator) Notify(notification *Notification) (slackThreadLink string) {
	if notification.SlackChannel == "" {
		notification.SlackChannel = DefaultSlackChannel
	}
	if n.slackClient == nil {
		fmt.Println("Slack channel is empty or client is not initialized")
		return ""
	}
	color, blocks := n.buildBlocks(notification)
	_ = color
	attachment := slack.Attachment{
		Color: color,
		Blocks: slack.Blocks{
			BlockSet: blocks,
		},
	}

	s1, s2, err := n.slackClient.PostMessage(notification.SlackChannel, slack.MsgOptionAttachments(attachment))
	if err != nil {
		fmt.Println("Cannot send msg to slack", err)
		return ""
	} else {
		log.Println("Slack msg sent successfully: ", s1, s2)
		return fmt.Sprintf("https://ocgwp.slack.com/archives/%v/p%v", s1, strings.Replace(s2, ".", "", -1))
	}
}

func (n *Notificator) buildBlocks(notification *Notification) (string, []slack.Block) {
	blocks := []slack.Block{}

	color := "#00ff00"
	allFailed := notification.Passed == 0
	isNoCaseFound := notification.Passed+notification.Failed+notification.TimedOut+notification.Skipped == 0
	comment := ""
	// build comment & color
	if isNoCaseFound {
		comment = "`No case found`"
		color = "#ff0000"
	} else if allFailed {
		comment = "`Whole tests failed`"
		color = "#ff0000"
	} else {
		if notification.Failed+notification.TimedOut == 0 && notification.Passed > 0 {
			color = "#6ac46a"
		} else if notification.Passed > notification.Failed+notification.TimedOut {
			color = "#0000ff"
		} else {
			color = "#ffff00"
		}
	}

	phongdoUserId := "UC0CE05JP"
	truongbuiUserId := "U0NCL405Q"
	users := fmt.Sprintf("<@%v> <@%v>", phongdoUserId, truongbuiUserId)
	if notification.SlackUserID != "" {
		users = fmt.Sprintf("<@%v>", notification.SlackUserID)
	}
	//if notification.SlackLeadID != "" && notification.SlackUserID != notification.SlackLeadID {
	//	users = fmt.Sprintf("%v <@%v>", users, notification.SlackLeadID)
	//}

	if notification.JobName == "" {
		notification.JobName = "unknown (an ad-hoc job)"
	}

	blocks = append(blocks, slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*Hi* %v, here is the *report* of job `(%v)` running against `(%v)` env. %v", users, notification.JobName, notification.Env, comment), false, false), nil, nil))

	blocks = append(blocks, slack.NewSectionBlock(nil, []*slack.TextBlockObject{
		slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*Passed:* %v", notification.Passed), false, false),
		slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*Failed:* %v", notification.Failed), false, false),
		slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*Timed Out:* %v", notification.TimedOut), false, false),
		slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*Skipped:* %v", notification.Skipped), false, false),
		slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*Run Mode:* %v", notification.RunMode), false, false),
		slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*Job Name:* %v", notification.JobName), false, false),
		slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*Running Time:* %v", time.Since(notification.StartedTime)), false, false),
	}, nil))

	testhubLink := ""
	if notification.ThJobID > 0 {
		testhubLink = fmt.Sprintf(`https://test-hub.ocg.to/admin/th/case/?jobs__pk__exact=%v`, notification.ThJobID)
	}
	runsLink := ""
	reportLink := ""
	if notification.ThRunGroupId > 0 {
		runsLink = fmt.Sprintf("https://test-hub.ocg.to/admin/th/run/?run_group__pk__exact=%v", notification.ThRunGroupId)
		reportLink = fmt.Sprintf("https://report-autopilot.shopbase.dev/show?rungroup_id=%v", notification.ThRunGroupId)
	}

	footer := []string{}
	if testhubLink != "" {
		footer = append(footer, fmt.Sprintf("<%v|List cases>", testhubLink))
	}
	if runsLink != "" {
		footer = append(footer, fmt.Sprintf("<%v|T.Hub RunGroup>", runsLink))
	}
	if reportLink != "" {
		footer = append(footer, fmt.Sprintf("<%v|Report of RunGroup>", reportLink))
	}
	if notification.BuildURL != "" {
		footer = append(footer, fmt.Sprintf("<%v|Jenkins Job>", notification.BuildURL))
	}

	blocks = append(blocks, slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", strings.Join(footer, " OR "), false, false), nil, nil))

	return color, blocks
}

func (n *Notificator) NotifySimple(channel string, message string) error {
	if channel == "" {
		channel = DefaultSlackChannel
	}
	if n.slackClient == nil {
		fmt.Println("Slack channel is empty or client is not initialized")
		return errors.New("Slack client is  is")
	}

	notifyBlock := slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", message, false, false), nil, nil)
	blocks := make([]slack.Block, 0)
	blocks = append(blocks, notifyBlock)

	attachment := slack.Attachment{
		Color: "#ff0000",
		Blocks: slack.Blocks{
			BlockSet: blocks,
		},
	}

	s1, s2, err := n.slackClient.PostMessage(channel, slack.MsgOptionAttachments(attachment))
	if err != nil {
		fmt.Println("Cannot send msg to slack", err)
		return err
	} else {
		log.Println(fmt.Sprintf("https://ocgwp.slack.com/archives/%v/p%v", s1, strings.Replace(s2, ".", "", -1)))
		return nil
	}
}
