package config

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) GetConfig(c *gin.Context) {
	c.JSON(http.StatusOK, AppConfig.Authentication)
}
