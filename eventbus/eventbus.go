package eventbus

import (
	"fmt"

	"github.com/hzwesoft-github/underscore/lang"
	"github.com/hzwesoft-github/underscore/openwrt"
	// "github.com/hzwesoft-github/underscore/openwrt"
)

type SendMode int

const (
	LOCAL SendMode = iota
	REMOTE
	BOTH
)

type Event struct {
	Topic   string
	Payload any
	Local   bool
	Remote  bool
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
	fmt.Printf("%v\n", subscribers)
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
	client, err := openwrt.NewUbusClient(true)
	if err != nil {
		return err
	}

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
