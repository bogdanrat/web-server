package models

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	QRCode   string `json:"qr_code"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type SignUpRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
