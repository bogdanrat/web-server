package queue

type EventEmitter interface {
	Emit(Event) error
}
