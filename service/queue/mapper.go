package queue

import (
	"errors"
)

type MapperType int

const (
	StaticMapper MapperType = iota
)

type EventMapper interface {
	MapEvent(eventName string, serialized interface{}) (Event, error)
}

func NewEventMapper(mapperType MapperType) (EventMapper, error) {
	switch mapperType {
	case StaticMapper:
		return &StaticEventMapper{}, nil
	default:
		return nil, errors.New("unknown event mapper type")
	}
}
