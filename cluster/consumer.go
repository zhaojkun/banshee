package cluster

import (
	"encoding/json"
	"time"

	"github.com/eleme/banshee/models"
	"github.com/eleme/banshee/storage"
	"github.com/eleme/banshee/util/log"
	"github.com/streadway/amqp"
)

type consumer struct {
	opts        *Options
	db          *storage.DB
	conn        *amqp.Connection
	channel     *amqp.Channel
	closeNotify chan *amqp.Error
}

func newConsumer(opts *Options, db *storage.DB) *consumer {
	c := &consumer{
		opts: opts,
		db:   db,
	}
	go c.watch()
	return c
}

func (c *consumer) watch() {
	for {
		msgs, err := c.reconnect()
		if err == nil {
			c.work(msgs)
		} else {
			log.Infof("Connect to server meets error:%v", err)
		}
		time.Sleep(30 * time.Millisecond)
	}
}

func (c *consumer) work(msgs <-chan amqp.Delivery) {
	for {
		select {
		case msg := <-msgs:
			var m models.Message
			err := json.Unmarshal(msg.Body, &m)
			if err != nil {
				continue
			}
			if m.Rule == nil {
				continue
			}
			log.Infof("received message type: %s,rule id:%d", m.Type, m.Rule.ID)
			if m.Type == models.RULEADD {
				c.db.Admin.RulesCache.Put(m.Rule)
			} else if m.Type == models.RULEDELETE {
				c.db.Admin.RulesCache.Delete(m.Rule.ID)
			}
		case <-c.closeNotify:
			return
		}
	}
}

func (c *consumer) reconnect() (<-chan amqp.Delivery, error) {
	if c.channel != nil {
		c.channel.Close()
		c.channel = nil
	}
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	conn, err := amqp.Dial(c.opts.DSN)
	if err != nil {
		return nil, err
	}
	c.closeNotify = conn.NotifyClose(make(chan *amqp.Error))
	c.conn = conn
	ch, err := c.conn.Channel()
	if err != nil {
		return nil, err
	}
	err = ch.ExchangeDeclare(c.opts.ExchangeName, ExchangeType, false, false, false, false, nil)
	if err != nil {
		ch.Close()
		return nil, err
	}
	q, err := ch.QueueDeclare(c.opts.QueueName, false, false, false, false, nil)
	if err != nil {
		ch.Close()
		return nil, err
	}
	err = ch.QueueBind(q.Name, "", c.opts.ExchangeName, false, nil)
	if err != nil {
		ch.Close()
		return nil, err
	}
	c.channel = ch
	return ch.Consume(q.Name, "", true, false, false, false, nil)
}
