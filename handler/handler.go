package handler

import (
	"github.com/bogdanrat/web-server/cache"
	"github.com/bogdanrat/web-server/repository"
	"github.com/gin-gonic/gin"
	"net/http"
)

var (
	repo        repository.DatabaseRepository
	cacheClient cache.Client
)

func InitRepository(dbRepo repository.DatabaseRepository) {
	repo = dbRepo
}

func InitCache(client cache.Client) {
	cacheClient = client
}

func GetCache() cache.Client {
	return cacheClient
}

func GetUsers(c *gin.Context) {
	users, err := repo.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, users)
}
