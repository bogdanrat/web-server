package router

import (
	"github.com/bogdanrat/web-server/cache"
	"github.com/bogdanrat/web-server/config"
	"github.com/bogdanrat/web-server/handler/authentication"
	"github.com/bogdanrat/web-server/handler/users"
	"github.com/bogdanrat/web-server/middleware"
	"github.com/bogdanrat/web-server/repository"
	pb "github.com/bogdanrat/web-server/service/auth/proto"
	"github.com/gin-gonic/gin"
	"net/http"
)

func New(repo repository.DatabaseRepository, cacheClient cache.Client, authClient pb.AuthClient) http.Handler {
	router := gin.Default()
	gin.SetMode(config.AppConfig.Server.GinMode)

	authenticationHandler := authentication.NewHandler(repo, cacheClient, authClient)
	usersHandler := users.NewHandler(repo)

	// public endpoints
	router.POST("/sign-up", authenticationHandler.SignUp)
	router.POST("/login", authenticationHandler.Login)
	router.POST("/logout", authenticationHandler.Logout)
	router.POST("/token/refresh", authenticationHandler.RefreshToken)

	// private endpoints, requires jwt
	protectedGroup := router.Group("/api").Use(middleware.Authorization(authenticationHandler.Cache, authenticationHandler.RPC))
	protectedGroup.GET("/users", usersHandler.GetUsers)

	return router
}
