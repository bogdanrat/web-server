package models

type UserSignUp struct {
	User    User   `json:"user"`
	QrImage []byte `json:"qr_code,omitempty"`
}

// Name returns the event's name
func (e *UserSignUp) Name() string {
	return "userSignUp"
}
