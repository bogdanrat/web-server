package auth

import (
	"errors"
	"github.com/bogdanrat/web-server/config"
	"github.com/bogdanrat/web-server/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/twinj/uuid"
	"time"
)

const (
	AccessSecretKey  = "verySecretKey"
	RefreshSecretKey = "anotherVerySecretKey"
	Issuer           = "AuthService"
)

type JwtAccessClaims struct {
	Email      string
	AccessUUID string
	jwt.StandardClaims
}

type JwtRefreshClaims struct {
	Email       string
	RefreshUUID string
	jwt.StandardClaims
}

// GenerateToken generates new JWT Access & Refresh tokens
func GenerateToken(email string) (*models.Token, error) {
	// generate access token
	accessClaims := &JwtAccessClaims{
		Email: email,
		// Since the UUID is unique each time it is created, a use can create more than one token.
		// This happens when a user is logged in on different devices.
		// The user can also logout from any of the devices without being logged out from all devices.
		AccessUUID: uuid.NewV4().String(),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Duration(config.AppConfig.Authentication.AccessTokenDuration) * time.Minute).Unix(),
			Issuer:    Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err := token.SignedString([]byte(AccessSecretKey))

	if err != nil {
		return nil, err
	}

	// generate refresh token
	refreshClaims := &JwtRefreshClaims{
		Email:       email,
		RefreshUUID: uuid.NewV4().String(),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Duration(config.AppConfig.Authentication.RefreshTokenDuration) * time.Minute).Unix(),
			Issuer:    Issuer,
		},
	}

	token = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err := token.SignedString([]byte(RefreshSecretKey))

	if err != nil {
		return nil, err
	}

	return &models.Token{
		AccessToken:         accessToken,
		AccessTokenExpires:  accessClaims.ExpiresAt,
		AccessUUID:          accessClaims.AccessUUID,
		RefreshToken:        refreshToken,
		RefreshTokenExpires: refreshClaims.ExpiresAt,
		RefreshUUID:         refreshClaims.RefreshUUID,
	}, nil
}

// ValidateAccessToken validates the JWT Access AccessToken
func ValidateAccessToken(signedToken string) (*JwtAccessClaims, error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JwtAccessClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(AccessSecretKey), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JwtAccessClaims)
	if !ok {
		return nil, errors.New("could not parse jwt access claim")
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		return nil, errors.New("jwt access token expired")
	}

	return claims, nil
}

func ValidateRefreshToken(signedToken string) (*JwtRefreshClaims, error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JwtRefreshClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(RefreshSecretKey), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JwtRefreshClaims)
	if !ok {
		return nil, errors.New("could not parse jwt refresh claim")
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		return nil, errors.New("jwt refresh token expired")
	}

	return claims, nil
}
