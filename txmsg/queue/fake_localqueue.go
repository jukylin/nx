package queue

import (
	entity "study-go/txmsg/domain/entity"
)

type FakeLocalQueue struct{}

func (flq FakeLocalQueue) Close() error {
	var r0 error

	return r0
}

func (flq FakeLocalQueue) Consumer(arg0 func(interface{})) bool {
	var r0 bool

	return r0
}

func (flq FakeLocalQueue) DroppedItem(arg0 interface{}) {

	return
}

func (flq FakeLocalQueue) Produce(arg0 entity.Msg) bool {
	var r0 bool
	if arg0.ID == 2 {
		r0 = true
	}
	return r0
}
