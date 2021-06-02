package router

import (
	pb "github.com/bogdanrat/web-server/contracts/proto/auth_service"
	"github.com/bogdanrat/web-server/contracts/proto/storage_service"
	"github.com/bogdanrat/web-server/service/core/cache"
	"github.com/bogdanrat/web-server/service/core/config"
	"github.com/bogdanrat/web-server/service/core/handler/authentication"
	"github.com/bogdanrat/web-server/service/core/handler/file"
	"github.com/bogdanrat/web-server/service/core/handler/users"
	"github.com/bogdanrat/web-server/service/core/middleware"
	"github.com/bogdanrat/web-server/service/core/repository"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
	"net/http"
)

func New(repo repository.DatabaseRepository, cacheClient cache.Client, authClient pb.AuthClient, storageClient storage_service.StorageClient) http.Handler {
	router := gin.Default()
	gin.SetMode(config.AppConfig.Server.GinMode)

	var options []grpc.CallOption

	// Compression: use network bandwidth efficiently when performing RPCs between client and services.
	// From the server side, registered compressors will be used automatically to decode request message and encode the responses.
	if config.AppConfig.Services.Auth.GRPC.UseCompression {
		options = append(options, grpc.UseCompressor(gzip.Name))
	}

	authenticationHandler := authentication.NewHandler(repo, cacheClient, &authentication.RPCConfig{
		Client:      authClient,
		Deadline:    config.AppConfig.Services.Auth.GRPC.Deadline,
		CallOptions: options,
	})

	usersHandler := users.NewHandler(repo)

	options = []grpc.CallOption{}
	if config.AppConfig.Services.Storage.GRPC.UseCompression {
		options = append(options, grpc.UseCompressor(gzip.Name))
	}

	fileHandler := file.NewHandler(&file.RPCConfig{
		Client:      storageClient,
		Deadline:    config.AppConfig.Services.Storage.GRPC.Deadline,
		CallOptions: options,
	})

	// public endpoints
	router.POST("/sign-up", authenticationHandler.SignUp)
	router.POST("/login", authenticationHandler.Login)
	router.POST("/logout", authenticationHandler.Logout)
	router.POST("/token/refresh", authenticationHandler.RefreshToken)

	// private endpoints, requires jwt
	protectedGroup := router.Group("/api").Use(middleware.Authorization(authenticationHandler.Cache, authenticationHandler.RPC.Client))
	protectedGroup.GET("/users", usersHandler.GetUsers)

	protectedGroup.POST("/files", fileHandler.PostFiles)
	protectedGroup.GET("/files", fileHandler.GetFiles)
	protectedGroup.DELETE("/file", fileHandler.DeleteFile)
	protectedGroup.DELETE("/files", fileHandler.DeleteFiles)

	return router
}
