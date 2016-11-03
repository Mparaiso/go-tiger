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

package signal

type Event interface{}

// Signal is an implementation of the signal pattern
type Signal interface {
	Add(Listener)
	Remove(Listener)
	Dispatch(data Event) error
}

// Listener handle signals
type Listener interface {
	Handle(data Event) error
}

type funcListener struct {
	Listener func(data Event) error
}

func (fl funcListener) Handle(event Event) error {
	return fl.Listener(event)
}

func ListenerFunc(f func(data Event) error) *funcListener {
	return &funcListener{f}
}

type DefaultSignal struct {
	Listeners []Listener
}

func NewDefaultSignal() *DefaultSignal {
	return &DefaultSignal{Listeners: []Listener{}}
}

func (signal *DefaultSignal) Add(l Listener) {
	if signal.IndexOf(l) != -1 {
		return
	}
	signal.Listeners = append(signal.Listeners, l)
}

func (signal *DefaultSignal) IndexOf(l Listener) int {
	for i, listener := range signal.Listeners {
		if listener == l {
			return i
		}
	}
	return -1
}

func (signal *DefaultSignal) Remove(l Listener) {
	index := signal.IndexOf(l)
	if index == -1 {
		return
	}
	head := signal.Listeners[:index]
	if index == len(signal.Listeners)-1 {
		signal.Listeners = head
	} else {
		signal.Listeners = append(head, signal.Listeners[index+1:]...)
	}
}

func (signal *DefaultSignal) Dispatch(data Event) error {
	for _, listener := range signal.Listeners {
		if err := listener.Handle(data); err != nil {
			return err
		}
	}
	return nil
}
