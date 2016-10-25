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

// Notify event
func (w *WebHook) Notify(hook models.WebHook, ew *models.EventWrapper) error {
	body, _ := json.Marshal(ew)
	req, err := http.NewRequest("POST", hook.URL, bytes.NewReader(body))
	req.Header.Add("Content-Type", "application/json")
	_, err = http.DefaultClient.Do(req)
	return err
}

func init() {
	alerter.RegisterNotifier("webhook", new(WebHook))
}
