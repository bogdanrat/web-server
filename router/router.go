package router

import (
	"github.com/bogdanrat/web-server/config"
	"github.com/bogdanrat/web-server/handler"
	"github.com/bogdanrat/web-server/middleware"
	"github.com/gin-gonic/gin"
	"net/http"
)

func New() http.Handler {
	router := gin.Default()
	gin.SetMode(config.AppConfig.Server.GinMode)

	// public endpoints
	router.POST("/sign-up", handler.SignUp)
	router.POST("/login", handler.Login)
	router.POST("/logout", handler.Logout)
	router.POST("/token/refresh", handler.RefreshToken)

	// private endpoints, requires jwt
	protectedGroup := router.Group("/api").Use(middleware.Authorization(handler.GetCache()))
	protectedGroup.GET("/users", handler.GetUsers)

	return router
}
