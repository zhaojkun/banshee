package webhook

import "github.com/eleme/banshee/models"

type WebHook interface {
	Delivery(ev *models.Event) error
}
