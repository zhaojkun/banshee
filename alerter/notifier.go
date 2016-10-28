package alerter

import "github.com/eleme/banshee/models"

// Notifiers is the collections of supported notifiers
var Notifiers = make(map[string]Notifier)

// Notifier should be implement to notify the events
type Notifier interface {
	Notify(hook models.WebHook, ew *models.EventWrapper) error
}

// RegisterNotifier into the Notifiers
func RegisterNotifier(name string, wk Notifier) {
	Notifiers[name] = wk
}
