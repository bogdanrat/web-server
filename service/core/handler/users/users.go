package users

import (
	"context"
	"github.com/bogdanrat/web-server/contracts/models"
	pb "github.com/bogdanrat/web-server/contracts/proto/database_service"
	"github.com/bogdanrat/web-server/service/core/lib"
	"github.com/bogdanrat/web-server/service/core/repository"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"net/http"
	"time"
)

type RPCConfig struct {
	Client      pb.DatabaseClient
	CallOptions []grpc.CallOption
	Deadline    int64
}

type Handler struct {
	Repository repository.DatabaseRepository
	RPC        *RPCConfig
}

func NewHandler(repo repository.DatabaseRepository, rpcConfig *RPCConfig) *Handler {
	return &Handler{
		Repository: repo,
		RPC:        rpcConfig,
	}
}

func (h *Handler) GetUsers(c *gin.Context) {
	deadline := time.Now().Add(time.Millisecond * time.Duration(h.RPC.Deadline))
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	response, err := h.RPC.Client.GetAllUsers(ctx, &pb.GetAllUsersRequest{})

	if err != nil {
		if jsonErr := lib.HandleRPCError(err); err != nil {
			c.JSON(jsonErr.StatusCode, jsonErr)
			return
		}
	}

	users := make([]*models.User, 0)

	for _, resUser := range response.Users {
		user := &models.User{
			Name:     resUser.GetName(),
			Email:    resUser.GetEmail(),
			Password: resUser.GetPassword(),
		}
		qrSecret := resUser.GetQrSecret()
		user.QRSecret = &qrSecret

		users = append(users, user)
	}

	c.JSON(http.StatusOK, users)
}
