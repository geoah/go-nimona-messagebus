package events

type EventBus interface {
	HandleEvent(handler func(ev *Event) error) error
	Send(ev *Event) error
}
