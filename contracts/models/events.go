package models

const (
	UserSignUpEventName         = "userSignUp"
	NewKeyValuePairEventName    = "newKeyValuePair"
	DeleteKeyValuePairEventName = "deleteKeyValuePair"
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
	Pairs []*KeyValuePair
}

func (e *NewKeyValuePairEvent) Name() string {
	return NewKeyValuePairEventName
}

type DeleteKeyValuePairEvent struct {
	Pair *KeyValuePair
}

func (e *DeleteKeyValuePairEvent) Name() string {
	return DeleteKeyValuePairEventName
}
