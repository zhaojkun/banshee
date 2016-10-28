package notifier

import "github.com/eleme/banshee/config"

// Globals
var (
	// Config
	cfg *config.Config
)

// Init Notifier
func Init(config *config.Config) {
	cfg = config
}
