package models

type UserSignUpEvent struct {
	User    *User  `json:"user"`
	QrImage []byte `json:"qr_code,omitempty"`
}

// Name returns the event's name
func (e *UserSignUpEvent) Name() string {
	return "userSignUp"
}
