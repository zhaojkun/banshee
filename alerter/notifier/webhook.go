package notifier

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/eleme/banshee/alerter"
	"github.com/eleme/banshee/models"
)

// WebHook notifier
type WebHook struct {
	client *http.Client
}

// NewWebHook create a webhook client for notification.
func NewWebHook() *WebHook {
	return &WebHook{
		client: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

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
	_, err = w.client.Do(req)
	return err
}

func init() {
	alerter.RegisterNotifier("webhook", NewWebHook())
}
