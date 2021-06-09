package queue

type EventListener interface {
	Listen(...string) (<-chan Event, <-chan error, error)
	EventMapper() EventMapper
}
