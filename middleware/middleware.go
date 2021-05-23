package middleware

import (
	"github.com/bogdanrat/web-server/auth"
	"github.com/bogdanrat/web-server/cache"
	"github.com/bogdanrat/web-server/models"
	"github.com/bogdanrat/web-server/util"
	"github.com/gin-gonic/gin"
)

// Authorization validates jwt and authorizes users based by Header 'Authorization Bearer {{token}}'
func Authorization(cacheClient cache.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, jsonErr := util.ExtractToken(c.Request)
		if jsonErr != nil {
			c.JSON(jsonErr.StatusCode, jsonErr)
			// Abort prevents pending handlers from being called.
			// If the authorization fails, call Abort to ensure the remaining handlers for this request are not called.
			c.Abort()
			return
		}

		claims, err := auth.ValidateAccessToken(token)
		if err != nil {
			jsonErr = models.NewUnauthorizedError(err.Error())
			c.JSON(jsonErr.StatusCode, jsonErr)
			c.Abort()
			return
		}

		_, err = cacheClient.Get(claims.AccessUUID)
		if err != nil {
			jsonErr = models.NewUnauthorizedError("authorization token expired")
			c.JSON(jsonErr.StatusCode, jsonErr)
			c.Abort()
			return
		}

		// It executes the pending handlers in the chain inside the calling handler
		c.Next()
	}
}
