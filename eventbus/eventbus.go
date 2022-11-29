package eventbus

import (
	"github.com/hzwesoft-github/underscore/lang"
	"github.com/hzwesoft-github/underscore/openwrt"
)

type Event struct {
	Topic       string
	Payload     any
	ContextData any
	Local       bool
	Remote      bool
}

type OnEvent func(Event) error

var (
	subscribers lang.SliceMap[string, OnEvent]
)

func init() {
	subscribers = lang.NewSliceMap[string, OnEvent]()
}

func Register(topic string, cb OnEvent) {
	lang.AddSliceMapValue(subscribers, topic, cb)
}

func SendLocal(event Event, async bool) error {
	callbacks, ok := subscribers[event.Topic]
	if !ok {
		return nil
	}

	if async {
		go func() {
			for _, cb := range callbacks {
				cb(event)
			}
		}()
	} else {
		for _, cb := range callbacks {
			if err := cb(event); err != nil {
				return err
			}
		}
	}

	return nil
}

func SendRemote(topic string, payload any) error {
	client, err := openwrt.NewUbusClient(false)
	if err != nil {
		return err
	}
	defer client.Free()

	return client.SendEvent(topic, payload)
}

func SendEvent(event Event, async bool) error {
	if event.Local {
		if err := SendLocal(event, async); err != nil {
			return err
		}
	}

	if event.Remote {
		if err := SendRemote(event.Topic, event.Payload); err != nil {
			return err
		}
	}

	return nil
}

func Validate(event Event) error {
	if err := SendLocal(event, false); err != nil {
		return err
	}

	return nil
}

func NewEvent(topic string, payload any, context any, local, remote bool) Event {
	return Event{
		Topic:       topic,
		Payload:     payload,
		ContextData: context,
		Local:       local,
		Remote:      remote,
	}
}

func NewLocalEvent(topic string, payload any, context any) Event {
	return Event{
		Topic:       topic,
		Payload:     payload,
		ContextData: context,
		Local:       true,
	}
}

func NewRemoteEvent(topic string, payload any, context any) Event {
	return Event{
		Topic:       topic,
		Payload:     payload,
		ContextData: context,
		Remote:      true,
	}
}
