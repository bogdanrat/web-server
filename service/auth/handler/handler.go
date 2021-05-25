package handler

import (
	"bytes"
	"context"
	"encoding/base32"
	"github.com/bogdanrat/web-server/service/auth/lib"
	pb "github.com/bogdanrat/web-server/service/auth/proto"
	"github.com/dgryski/dgoogauth"
	"image/png"
)

type AuthServer struct{}

func (s *AuthServer) GenerateQRCode(ctx context.Context, req *pb.GenerateQRCodeRequest) (*pb.GenerateQRCodeResponse, error) {
	img, secret, err := lib.GenerateQRCode(req.Email)
	if err != nil {
		return nil, err
	}

	buffer := &bytes.Buffer{}
	if err := png.Encode(buffer, img); err != nil {
		return nil, err
	}

	return &pb.GenerateQRCodeResponse{
		Image:  buffer.Bytes(),
		Secret: secret,
	}, nil
}

func (s *AuthServer) ValidateQRCode(ctx context.Context, req *pb.ValidateQRCodeRequest) (*pb.ValidateQRCodeResponse, error) {
	encodedSecret := base32.StdEncoding.EncodeToString([]byte(req.QrSecret))

	otpc := &dgoogauth.OTPConfig{
		Secret:      encodedSecret,
		WindowSize:  0,
		HotpCounter: 0,
	}

	authenticated, err := otpc.Authenticate(req.QrCode)
	if err != nil {
		return &pb.ValidateQRCodeResponse{Validated: false}, err
	}

	return &pb.ValidateQRCodeResponse{Validated: authenticated}, nil
}

func (s *AuthServer) GenerateToken(ctx context.Context, req *pb.GenerateTokenRequest) (*pb.GenerateTokenResponse, error) {
	token, err := lib.GenerateToken(req.Email, req.AccessTokenDuration, req.RefreshTokenDuration)
	if err != nil {
		return nil, err
	}

	return &pb.GenerateTokenResponse{
		Token: token,
	}, nil
}

func (s *AuthServer) ValidateAccessToken(ctx context.Context, req *pb.ValidateAccessTokenRequest) (*pb.ValidateAccessTokenResponse, error) {
	claims, err := lib.ValidateAccessToken(req.SignedToken)
	if err != nil {
		return nil, err
	}

	return &pb.ValidateAccessTokenResponse{
		Email:      claims.Email,
		AccessUuid: claims.AccessUUID,
	}, nil
}

func (s *AuthServer) ValidateRefreshToken(ctx context.Context, req *pb.ValidateRefreshTokenRequest) (*pb.ValidateRefreshTokenResponse, error) {
	claims, err := lib.ValidateRefreshToken(req.SignedToken)
	if err != nil {
		return nil, err
	}

	return &pb.ValidateRefreshTokenResponse{
		Email:       claims.Email,
		RefreshUuid: claims.RefreshUUID,
	}, nil
}
