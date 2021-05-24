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
