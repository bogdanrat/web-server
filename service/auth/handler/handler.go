package handler

import (
	"context"
	pb "github.com/bogdanrat/web-server/service/auth/proto"
)

type AuthServer struct{}

func (s *AuthServer) SignUp(context.Context, *pb.SignUpRequest) (*pb.SignUpResponse, error) {
	return nil, nil
}
