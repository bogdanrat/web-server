package middleware

import (
	"context"
	"github.com/bogdanrat/web-server/contracts/models"
	pb "github.com/bogdanrat/web-server/contracts/proto/auth_service"
	"github.com/bogdanrat/web-server/service/core/cache"
	"github.com/bogdanrat/web-server/service/core/lib"
	"github.com/bogdanrat/web-server/service/core/util"
	"github.com/gin-gonic/gin"
	"strings"
)

var (
	pathsToSkipFromAuthorization = []string{"/sign-up", "/login", "/logout", "/token/refresh"}
)

// Authorization validates jwt and authorizes users based by Header 'Authorization Bearer {{token}}'
func Authorization(developmentMode bool, cacheClient cache.Client, authClient pb.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		if developmentMode == false {
			token, jsonErr := util.ExtractToken(c.Request)
			if jsonErr != nil {
				c.JSON(jsonErr.StatusCode, jsonErr)
				// Abort prevents pending handlers from being called.
				// If the authorization fails, call Abort to ensure the remaining handlers for this request are not called.
				c.Abort()
				return
			}

			response, err := authClient.ValidateAccessToken(context.Background(), &pb.ValidateAccessTokenRequest{SignedToken: token})
			if jsonErr = lib.HandleRPCError(err); jsonErr != nil {
				c.JSON(jsonErr.StatusCode, jsonErr)
				c.Abort()
				return
			}

			_, err = cacheClient.Get(response.AccessUuid)
			if err != nil {
				jsonErr = models.NewUnauthorizedError("authorization token expired")
				c.JSON(jsonErr.StatusCode, jsonErr)
				c.Abort()
				return
			}
		}

		// It executes the pending handlers in the chain inside the calling handler
		c.Next()
	}
}

func shouldSkipPath(path string) bool {
	for _, pathToSkip := range pathsToSkipFromAuthorization {
		if strings.Contains(path, pathToSkip) {
			return true
		}
	}
	return false
}
