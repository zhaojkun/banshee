package cluster

import (
	"encoding/json"
	"time"

	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/storage"
	"github.com/eleme/banshee/util/log"
	"github.com/streadway/amqp"
)

type publisher struct {
	opts        *Options
	db          *storage.DB
	msgCh       chan *models.Message
	conn        *amqp.Connection
	channel     *amqp.Channel
	closeNotify chan *amqp.Error
}

func newPublisher(opts *Options, db *storage.DB) *publisher {
	p := &publisher{
		opts:  opts,
		db:    db,
		msgCh: make(chan *models.Message, bufferedChangedRulesLimit*2),
	}
	p.initRuleListener()
	go p.watch()
	return p
}

func (p *publisher) initRuleListener() {
	p.db.Admin.RulesCache.OnChange(p.msgCh)
}

func (p *publisher) watch() {
	for {
		if err := p.reconnect(); err == nil {
			p.work()
		} else {
			log.Infof("Connect to server meets error:%v", err)
		}
		time.Sleep(30 * time.Millisecond)
	}
}

func (p *publisher) work() {
	for {
		select {
		case msg := <-p.msgCh:
			err := p.publish(msg)
			if err != nil {
				return
			}
			log.Infof("publish message type: %s,rule id:%d", msg.Type, msg.Rule.ID)
		case <-p.closeNotify:
			return
		}
	}
}

func (p *publisher) reconnect() error {
	if p.channel != nil {
		p.channel.Close()
		p.channel = nil
	}
	if p.conn != nil {
		p.conn.Close()
		p.channel = nil
	}
	conn, err := amqp.Dial(p.opts.DSN)
	if err != nil {
		return err
	}
	p.closeNotify = conn.NotifyClose(make(chan *amqp.Error))
	p.conn = conn
	ch, err := p.conn.Channel()
	if err != nil {
		return err
	}
	err = ch.ExchangeDeclare(p.opts.ExchangeName, ExchangeType, false, false, false, false, nil)
	if err != nil {
		ch.Close()
		return err
	}
	p.channel = ch
	return nil
}

func (p *publisher) publish(msg *models.Message) error {
	if p.channel == nil {
		return amqp.ErrClosed
	}
	buf, err := json.Marshal(msg)
	if err != nil {
		return nil
	}
	return p.channel.Publish(p.opts.ExchangeName, "", false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        buf,
	})
}
