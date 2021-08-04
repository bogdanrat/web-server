package models

const (
	UserSignUpEventName      = "userSignUp"
	NewKeyValuePairEventName = "newKeyValuePair"
)

type UserSignUpEvent struct {
	User    *User  `json:"user"`
	QrImage []byte `json:"qr_code,omitempty"`
}

// Name returns the event's name
func (e *UserSignUpEvent) Name() string {
	return UserSignUpEventName
}

type NewKeyValuePairEvent struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

func (e *NewKeyValuePairEvent) Name() string {
	return NewKeyValuePairEventName
}
