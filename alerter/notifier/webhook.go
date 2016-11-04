package notifier

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/eleme/banshee/alerter"
	"github.com/eleme/banshee/models"
)

// WebHook notifier
type WebHook struct{}

// Event is the webhook payload
type Event struct {
	ID      string          `json:"id"`
	Comment string          `json:"comment"`
	Metric  *models.Metric  `json:"metric"`
	Rule    *models.Rule    `json:"rule"`
	Project *models.Project `json:"project"`
	Team    *models.Team    `json:"team"`
}

// Notify event
func (w *WebHook) Notify(hook models.WebHook, ew *models.EventWrapper) error {
	evt := Event{}
	evt.ID = ew.ID
	evt.Comment = ew.RuleTranslatedComment
	evt.Metric = ew.Metric
	evt.Rule = ew.Rule
	evt.Team = ew.Team
	evt.Project = ew.Project
	body, _ := json.Marshal(evt)
	req, err := http.NewRequest("POST", hook.URL, bytes.NewReader(body))
	req.Header.Add("Content-Type", "application/json")
	_, err = http.DefaultClient.Do(req)
	return err
}

func init() {
	alerter.RegisterNotifier("webhook", new(WebHook))
}
