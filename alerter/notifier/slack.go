package notifier

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eleme/banshee/alerter"
	"github.com/eleme/banshee/models"
)

// Slack Notifier
type Slack struct {
	name string
}

// Notify metric event
func (s *Slack) Notify(hook models.WebHook, ew *models.EventWrapper) error {
	text := packMessage(ew)
	graphiteMetricName := graphiteName(ew.Metric.Name)
	graphiteURL := getGrafanaPanelURL(graphiteMetricName)
	ruleURL := getRuleURL(ew.Project.TeamID, ew.Project.ID, ew.Rule.ID)
	grafanaLinkText := fmt.Sprintf("<%s|grafana图标>", graphiteURL)
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
		"attachments": []map[string]interface{}{
			{
				"title": title,
				"text":  content,
				"color": s.getColor(ew.Project.Name),
			},
		},
	}
	body, _ := json.Marshal(data)
	req, err := http.NewRequest("POST", hook.URL, bytes.NewReader(body))
	req.Header.Add("Content-Type", "application/json")
	_, err = http.DefaultClient.Do(req)
	return err
}

func (s *Slack) getColor(name string) string {
	hasher := md5.New()
	hasher.Write([]byte(name))
	encoded := hasher.Sum(nil)
	return "#" + hex.EncodeToString(encoded[:3])
}

func init() {
	alerter.RegisterNotifier("slack", &Slack{name: "banshee-bot"})
}
