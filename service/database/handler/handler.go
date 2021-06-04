package handler

import (
	"context"
	pb "github.com/bogdanrat/web-server/contracts/proto/database_service"
	"github.com/bogdanrat/web-server/service/database/db"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

type DatabaseServer struct {
	Database db.UsersDatabase
}

func New(database db.UsersDatabase) *DatabaseServer {
	return &DatabaseServer{
		Database: database,
	}
}

func (s *DatabaseServer) GetAllUsers(ctx context.Context, req *pb.GetAllUsersRequest) (*pb.GetAllUsersResponse, error) {
	users, err := s.Database.GetAllUsers()
	if err != nil {
		return nil, logError(status.Errorf(codes.Internal, "cannot get users: %v", err))
	}

	if err := contextError(ctx); err != nil {
		return nil, logError(err)
	}

	responseUsers := make([]*pb.User, 0)
	for _, user := range users {
		responseUser := &pb.User{
			Name:     user.Name,
			Email:    user.Email,
			Password: user.Password,
		}
		if user.QRSecret != nil {
			responseUser.QrSecret = *user.QRSecret
		}

		responseUsers = append(responseUsers, responseUser)
	}

	response := &pb.GetAllUsersResponse{
		Users: responseUsers,
	}

	return response, nil
}

func contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		return status.Errorf(codes.Canceled, "request was canceled by the client")
	case context.DeadlineExceeded:
		return status.Errorf(codes.DeadlineExceeded, "request deadline was exceeded")
	}
	return nil
}

func logError(err error) error {
	if err != nil {
		log.Println(err)
	}
	return err
}
