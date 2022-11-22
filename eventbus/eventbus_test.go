package eventbus

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

type Payload struct {
	Field int
}

func TestSyncEvent(t *testing.T) {
	Register("sync_event", Callback)

	SendEvent(Event{
		Topic:   "sync_event",
		Payload: Payload{1},
		Local:   true,
	}, false)

	Register("async_event", Callback)

	for i := 0; i < 1000; i++ {
		SendEvent(Event{
			Topic:   "async_event",
			Payload: Payload{i},
			Local:   true,
		}, true)
	}

	<-time.After(2 * time.Second)

	fmt.Println(counter)
}

var (
	counter int
	mutex   sync.Mutex
)

func Callback(event Event) error {
	fmt.Printf("%s callback invoked, payload %v\n", event.Topic, event.Payload)

	mutex.Lock()
	counter++
	defer mutex.Unlock()

	return nil
}
