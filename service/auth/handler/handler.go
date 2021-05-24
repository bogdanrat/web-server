package handler

import (
	"context"
	pb "github.com/bogdanrat/web-server/service/auth/proto"
)

type AuthServer struct{}

func (s *AuthServer) GenerateQRCode(context.Context, *pb.GenerateQRCodeRequest) (*pb.GenerateQRCodeResponse, error) {
	return nil, nil
}
