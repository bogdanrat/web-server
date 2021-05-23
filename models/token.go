package models

type Token struct {
	AccessToken        string
	AccessTokenExpires int64
	AccessUUID         string
	// RefreshToken is used to create new pairs of access and refresh tokens,
	// typically when the access token has expired, so the user does not have to login in again.
	RefreshToken        string
	RefreshTokenExpires int64
	RefreshUUID         string
}
