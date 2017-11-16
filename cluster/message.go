package cluster

import "github.com/eleme/banshee/storage"

// Exchanges
const (
	ExchangeType = "fanout"
)

const bufferedChangedRulesLimit = 128

// Options for message hub.
type Options struct {
	Master       bool   `json:"master"`
	DSN          string `json:"dsn"`
	VHost        string `json:"vHost"`
	ExchangeName string `json:"exchangeName"`
	QueueName    string `json:"queueName"`
}

// Hub is the message hub for rule changes.
type Hub struct {
	opts *Options
	db   *storage.DB
	pub  *publisher
	sub  *consumer
}

// New create a  Hub.
func New(opts *Options, db *storage.DB) (*Hub, error) {
	h := &Hub{
		opts: opts,
		db:   db,
	}
	if opts.Master {
		h.pub = newPublisher(opts, db)
	} else {
		h.sub = newConsumer(opts, db)
	}
	return h, nil
}
