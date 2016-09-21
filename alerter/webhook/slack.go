package webhook

import "github.com/eleme/banshee/models"

type Slack struct {
}

func (s *Slack) Delivery(ev *models.Event) error {
	mTitle := ev.Metric.Name
	if len(ev.Rule.Comment) != 0 {

	}
	return nil
}
