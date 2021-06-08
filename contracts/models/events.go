package models

type UserSignUp struct {
	User    User   `json:"user"`
	QrImage []byte `json:"qr_code,omitempty"`
}

// EventName returns the event's name
func (e *UserSignUp) EventName() string {
	return "userSignUp"
}
