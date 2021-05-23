package util

import (
	"github.com/bogdanrat/web-server/models"
	"net/http"
	"strings"
)

func ExtractToken(r *http.Request) (string, *models.JSONError) {
	authorization := r.Header.Get("Authorization")
	if authorization == "" {
		return "", models.NewUnauthorizedError("authorization header not found")
	}

	bearerToken := strings.Split(authorization, "Bearer ")
	if len(bearerToken) != 2 {
		return "", models.NewBadRequestError("malformed authorization token")
	}

	return strings.TrimSpace(bearerToken[1]), nil
}
