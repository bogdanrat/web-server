package queue

import (
	"encoding/json"
	"fmt"
	"github.com/bogdanrat/web-server/contracts/models"
	"github.com/mitchellh/mapstructure"
)

type StaticEventMapper struct{}

func (m *StaticEventMapper) MapEvent(eventName string, serialized interface{}) (Event, error) {
	var event Event

	switch eventName {
	case models.UserSignUpEventName:
		event = &models.UserSignUpEvent{}
	default:
		return nil, fmt.Errorf("unknown event type: %s", eventName)
	}

	switch s := serialized.(type) {
	case []byte:
		err := json.Unmarshal(s, event)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal event %s: %s", eventName, err)
		}
	default:
		cfg := &mapstructure.DecoderConfig{
			Result:  event,
			TagName: "json",
		}
		decoder, err := mapstructure.NewDecoder(cfg)
		if err != nil {
			return nil, fmt.Errorf("could not initialize decoder for event %s: %s", eventName, err)
		}

		err = decoder.Decode(s)
		if err != nil {
			return nil, fmt.Errorf("could not decode event %s: %s", eventName, err)
		}
	}

	return event, nil
}
