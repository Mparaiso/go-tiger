//    Copyright (C) 2016  mparaiso <mparaiso@online.fr>
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at

//      http://www.apache.org/licenses/LICENSE-2.0

//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package signal_test

import (
	"fmt"

	signal "github.com/Mparaiso/go-tiger/signal"
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
