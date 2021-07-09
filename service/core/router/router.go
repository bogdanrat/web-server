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
	"github.com/bogdanrat/web-server/service/queue"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
	"net/http"
)

func New(repo repository.DatabaseRepository, cacheClient cache.Client, authClient pb.AuthClient, storageClient storage_service.StorageClient, eventEmitter queue.EventEmitter) http.Handler {
	router := gin.Default()
	gin.SetMode(config.AppConfig.Server.GinMode)
	router.Use(cors.Default())

	var authOptions []grpc.CallOption

	// Compression: use network bandwidth efficiently when performing RPCs between client and services.
	// From the server side, registered compressors will be used automatically to decode request message and encode the responses.
	if config.AppConfig.Services.Auth.GRPC.UseCompression {
		authOptions = append(authOptions, grpc.UseCompressor(gzip.Name))
	}

	authenticationHandler := authentication.NewHandler(
		repo,
		cacheClient,
		&authentication.RPCConfig{
			Client:      authClient,
			Deadline:    config.AppConfig.Services.Auth.GRPC.Deadline,
			CallOptions: authOptions,
		},
		eventEmitter,
	)

	usersHandler := users.NewHandler(repo)

	authOptions = []grpc.CallOption{}
	if config.AppConfig.Services.Storage.GRPC.UseCompression {
		authOptions = append(authOptions, grpc.UseCompressor(gzip.Name))
	}

	fileHandler := file.NewHandler(&file.RPCConfig{
		Client:      storageClient,
		Deadline:    config.AppConfig.Services.Storage.GRPC.Deadline,
		CallOptions: authOptions,
	})

	// public endpoints
	router.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
		c.String(http.StatusOK, "App is alive!\n")
	})
	router.GET("/login", authenticationHandler.ShowLogin)

	router.POST("/sign-up", authenticationHandler.SignUp)
	router.POST("/login", authenticationHandler.Login)
	router.POST("/logout", authenticationHandler.Logout)
	router.POST("/token/refresh", authenticationHandler.RefreshToken)

	// private endpoints, requires jwt
	apiGroup := router.Group("/api").Use(middleware.Authorization(config.AppConfig.Server.DevelopmentMode, authenticationHandler.Cache, authenticationHandler.AuthService.Client))
	apiGroup.GET("/users", usersHandler.GetUsers)

	apiGroup.GET("/file-page", fileHandler.GetFilePage)
	apiGroup.GET("/file", fileHandler.GetFile)
	apiGroup.GET("/files", fileHandler.GetFiles)
	apiGroup.POST("/files", fileHandler.PostFiles)
	apiGroup.DELETE("/file", fileHandler.DeleteFile)
	apiGroup.DELETE("/files", fileHandler.DeleteFiles)

	return router
}
