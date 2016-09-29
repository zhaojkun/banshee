package bus

import (
	"reflect"
	"sync"
)

type HandlerFunc interface{}

type Bus struct {
	sync.RWMutex
	handlers map[string][]HandlerFunc
}

func New() *Bus {
	return &Bus{
		handlers: make(map[string][]HandlerFunc),
	}
}

func (b *Bus) Subscribe(topic string, h HandlerFunc) error {
	b.Lock()
	defer b.Unlock()
	handlerType := reflect.TypeOf(h)
	if handlerType.Kind() != reflect.Func {
		return nil
	}
	b.handlers[topic] = append(b.handlers[topic], h)
	return nil
}

func (b *Bus) UnSubscribe(topic string, handler HandlerFunc) {
	b.Lock()
	defer b.Unlock()
	if handlers, ok := b.handlers[topic]; ok {
		for idx, h := range handlers {
			if reflect.ValueOf(h) == reflect.ValueOf(handler) {
				b.handlers[topic] = append(b.handlers[topic][:idx], b.handlers[topic][idx+1:]...)
			}
		}
	}
}

func (b *Bus) Publish(topic string, args ...interface{}) {
	b.RLock()
	defer b.RUnlock()
	if handlers, ok := b.handlers[topic]; ok {
		for _, handler := range handlers {
			b.doPublish(handler, topic, args...)
		}
	}
}

func (b *Bus) doPublish(h HandlerFunc, topic string, args ...interface{}) {
	// add workers
	defer func() {
		// process panic
	}()
	passedArgs := make([]reflect.Value, 0)
	for _, arg := range args {
		passedArgs = append(passedArgs, reflect.ValueOf(arg))
	}
	reflect.ValueOf(h).Call(passedArgs)
}

var defaultBus = New()

func Subscribe(topic string, h HandlerFunc) error {
	return defaultBus.Subscribe(topic, h)
}

func UnSubscribe(topic string, h HandlerFunc) {
	defaultBus.UnSubscribe(topic, h)
}

func Publish(topic string, args ...interface{}) {
	defaultBus.Publish(topic, args...)
}
