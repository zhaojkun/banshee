package notifier

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/eleme/banshee/alerter"
	"github.com/eleme/banshee/models"
)

// Slack Notifier
type Slack struct {
	name   string
	client *http.Client
}

// NewSlack create a slack client for notification.
func NewSlack(name string) *Slack {
	return &Slack{
		name: name,
		client: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

// Notify metric event
func (s *Slack) Notify(hook models.WebHook, ew *models.EventWrapper) error {
	text := packMessage(ew)
	graphiteMetricName := graphiteName(ew.Metric.Name)
	graphiteURL := getGrafanaPanelURL(graphiteMetricName)
	ruleURL := getRuleURL(ew.Project.TeamID, ew.Project.ID, ew.Rule.ID)
	grafanaLinkText := fmt.Sprintf("<%s|grafana图表>", graphiteURL)
	ruleLinkText := fmt.Sprintf("<%s|调整规则>", ruleURL)
	content := fmt.Sprintf("%s %s %s", text, grafanaLinkText, ruleLinkText)
	mTitle := ew.Metric.Name
	ruleComment := ew.Rule.Comment
	if ruleComment != "" {
		mTitle = translateComment(ew.Rule.Pattern, ew.Metric.Name, ruleComment)
	}
	title := fmt.Sprintf("%s - %s", ew.Project.Name, mTitle)
	data := map[string]interface{}{
		"username": s.name,
		"channel":  hook.URL,
		"attachments": []map[string]interface{}{
			{
				"title": title,
				"text":  content,
				"color": s.getColor(ew.Project.Name),
			},
		},
	}
	body, _ := json.Marshal(data)
	req, err := http.NewRequest("POST", cfg.Notifier.SlackURL, bytes.NewReader(body))
	req.Header.Add("Content-Type", "application/json")
	_, err = s.client.Do(req)
	return err
}

func (s *Slack) getColor(name string) string {
	hasher := md5.New()
	hasher.Write([]byte(name))
	encoded := hasher.Sum(nil)
	return "#" + hex.EncodeToString(encoded[:3])
}

func init() {
	alerter.RegisterNotifier("slack", NewSlack("banshee-bot"))
}
