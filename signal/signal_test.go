package signal_test

import (
	"fmt"

	signal "github.com/Mparaiso/tiger-go-framework/signal"
)

func ExampleSignal() {
	type DummyEvent struct {
		Value string
	}
	s := signal.NewDefaultSignal()
	listener := signal.ListenerFunc(func(e signal.Event) error {
		switch event := e.(type) {
		case DummyEvent:
			fmt.Print(event.Value)
		}
		return nil
	})
	s.Add(listener)
	s.Dispatch(DummyEvent{Value: "Hello from signal"})
	s.Remove(listener)
	s.Dispatch(DummyEvent{Value: "This event will not be handled by any listener"})
	// Output:
	// Hello from signal
}
