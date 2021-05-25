package lib

import (
	pb "github.com/bogdanrat/web-server/service/auth/proto"
	"github.com/dgrijalva/jwt-go"
	"github.com/twinj/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
func GenerateToken(email string, accessTokenDuration int64, refreshTokenDuration int64) (*pb.Token, error) {
	// generate access token
	accessClaims := &JwtAccessClaims{
		Email: email,
		// Since the UUID is unique each time it is created, a use can create more than one token.
		// This happens when a user is logged in on different devices.
		// The user can also logout from any of the devices without being logged out from all devices.
		AccessUUID: uuid.NewV4().String(),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Duration(accessTokenDuration) * time.Minute).Unix(),
			Issuer:    Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err := token.SignedString([]byte(AccessSecretKey))

	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not sign access token")
	}

	// generate refresh token
	refreshClaims := &JwtRefreshClaims{
		Email:       email,
		RefreshUUID: uuid.NewV4().String(),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Duration(refreshTokenDuration) * time.Minute).Unix(),
			Issuer:    Issuer,
		},
	}

	token = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err := token.SignedString([]byte(RefreshSecretKey))

	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not sign refresh token")
	}

	return &pb.Token{
		AccessToken:         accessToken,
		AccessTokenExpires:  accessClaims.ExpiresAt,
		AccessUuid:          accessClaims.AccessUUID,
		RefreshToken:        refreshToken,
		RefreshTokenExpires: refreshClaims.ExpiresAt,
		RefreshUuid:         refreshClaims.RefreshUUID,
	}, status.New(codes.OK, "").Err()
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid token format")
	}

	claims, ok := token.Claims.(*JwtAccessClaims)
	if !ok {
		return nil, status.Errorf(codes.Internal, "could not parse jwt access claim")
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		return nil, status.Errorf(codes.PermissionDenied, "jwt access token expired")
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid token format")
	}

	claims, ok := token.Claims.(*JwtRefreshClaims)
	if !ok {
		return nil, status.Errorf(codes.Internal, "could not parse jwt refresh claim")
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		return nil, status.Errorf(codes.PermissionDenied, "jwt refresh token expired")
	}

	return claims, nil
}
