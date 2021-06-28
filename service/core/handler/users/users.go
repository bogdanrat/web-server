package users

import (
	"github.com/bogdanrat/web-server/service/core/lib"
	"github.com/bogdanrat/web-server/service/core/repository"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Handler struct {
	Repository repository.DatabaseRepository
}

func NewHandler(repo repository.DatabaseRepository) *Handler {
	return &Handler{
		Repository: repo,
	}
}

func (h *Handler) GetUsers(c *gin.Context) {
	users, err := h.Repository.GetAllUsers()
	if err != nil {
		if jsonErr := lib.HandleRPCError(err); err != nil {
			c.JSON(jsonErr.StatusCode, jsonErr)
			return
		}
	}

	c.JSON(http.StatusOK, users)
}
