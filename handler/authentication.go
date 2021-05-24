package handler

import (
	"bytes"
	"fmt"
	"github.com/bogdanrat/web-server/auth"
	"github.com/bogdanrat/web-server/cache"
	"github.com/bogdanrat/web-server/config"
	"github.com/bogdanrat/web-server/models"
	"github.com/bogdanrat/web-server/util"
	"github.com/gin-gonic/gin"
	"image/png"
	"net/http"
	"reflect"
	"strconv"
	"time"
)

func SignUp(c *gin.Context) {
	request := &models.SignUpRequest{}
	if err := c.ShouldBindJSON(request); err != nil {
		jsonErr := models.NewBadRequestError("invalid sign up request")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	if !util.IsValidEmail(request.Email) {
		jsonErr := models.NewBadRequestError("invalid email")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	if _, err := repo.GetUserByEmail(request.Email); err == nil {
		jsonErr := models.NewBadRequestError("email already registered")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	user := &models.User{
		Name:  request.Name,
		Email: request.Email,
	}

	if err := user.HashPassword(request.Password); err != nil {
		jsonErr := models.NewInternalServerError("invalid sign up request")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	if config.AppConfig.Authentication.MFA {
		img, secret, err := auth.GenerateQRCode(request.Email)
		if err != nil {
			jsonErr := models.NewInternalServerError("unable generate qr code")
			c.JSON(jsonErr.StatusCode, jsonErr)
			return
		}

		user.QRSecret = &secret

		buffer := &bytes.Buffer{}
		if err := png.Encode(buffer, img); err != nil {
			jsonErr := models.NewInternalServerError("unable to encode qr image")
			c.JSON(jsonErr.StatusCode, jsonErr)
			return
		}

		c.Writer.Header().Set("Content-Type", "image/png")
		c.Writer.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
		if _, err = c.Writer.Write(buffer.Bytes()); err != nil {
			c.JSON(http.StatusInternalServerError, "unable to write image")
			return
		}
	}

	if err := repo.InsertUser(user); err != nil {
		jsonErr := models.NewInternalServerError("unable to insert user into db")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	if reflect.ValueOf(cacheClient).Elem().Type().AssignableTo(reflect.TypeOf(cache.Redis{})) {
		cacheClient.(*cache.Redis).Publish(config.AppConfig.Authentication.Channel, user.Email)
	}
}

func Login(c *gin.Context) {
	request := &models.LoginRequest{}
	if err := c.ShouldBindJSON(request); err != nil {
		jsonErr := models.NewBadRequestError("invalid login request")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	user, err := repo.GetUserByEmail(request.Email)
	if err != nil {
		jsonErr := models.NewNotFoundError(fmt.Sprintf("user with email %s not found", request.Email))
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	if config.AppConfig.Authentication.MFA {
		if user.QRSecret == nil {
			jsonErr := models.NewUnprocessableEntityError("missing qr code when mfa is enabled")
			c.JSON(jsonErr.StatusCode, jsonErr)
			return
		}

		authenticated, err := auth.ValidateQRCode(request.QRCode, *user.QRSecret)
		if err != nil {
			jsonErr := models.NewBadRequestError("invalid qr code format")
			c.JSON(jsonErr.StatusCode, jsonErr)
			return
		}

		if !authenticated {
			jsonErr := models.NewUnauthorizedError("invalid qr code")
			c.JSON(jsonErr.StatusCode, jsonErr)
			return
		}
	}

	err = user.CheckPassword(request.Password)
	if err != nil {
		jsonErr := models.NewUnauthorizedError("invalid user credentials")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	token, err := auth.GenerateToken(request.Email)
	if err != nil {
		jsonErr := models.NewInternalServerError("could not sign token")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	if err = cacheClient.Set(token.AccessUUID, user.ID, int(time.Unix(token.AccessTokenExpires, 0).Sub(time.Now()).Seconds())); err != nil {
		jsonErr := models.NewInternalServerError("could not update cache")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	if err = cacheClient.Set(token.RefreshUUID, user.ID, int(time.Unix(token.RefreshTokenExpires, 0).Sub(time.Now()).Seconds())); err != nil {
		jsonErr := models.NewInternalServerError("could not update cache")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	tokenResponse := &models.LoginResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}
	c.JSON(http.StatusOK, tokenResponse)
}

func Logout(c *gin.Context) {
	token, jsonErr := util.ExtractToken(c.Request)
	if jsonErr != nil {
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	claims, err := auth.ValidateAccessToken(token)
	if err != nil {
		jsonErr = models.NewUnauthorizedError(err.Error())
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	_, err = cacheClient.Get(claims.AccessUUID)
	if err != nil {
		jsonErr = models.NewAlreadyReportedError("not logged in")
		c.JSON(jsonErr.StatusCode, jsonErr)
		c.Abort()
		return
	}

	err = cacheClient.Delete(claims.AccessUUID)
	if err != nil {
		jsonErr = models.NewInternalServerError("could not delete access token from cache")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	c.JSON(http.StatusOK, "Successfully logged out")
}

func RefreshToken(c *gin.Context) {
	mapToken := make(map[string]string, 0)
	if err := c.ShouldBindJSON(&mapToken); err != nil {
		jsonErr := models.NewBadRequestError("token malformed")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	refreshToken := mapToken["refresh_token"]
	claims, err := auth.ValidateRefreshToken(refreshToken)
	if err != nil {
		jsonErr := models.NewUnauthorizedError(err.Error())
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	// delete the previous refresh token
	err = cacheClient.Delete(claims.RefreshUUID)
	if err != nil {
		jsonErr := models.NewInternalServerError("could not delete refresh token from cache")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	// create new pairs of access & refresh tokens
	token, err := auth.GenerateToken(claims.Email)
	if err != nil {
		jsonErr := models.NewInternalServerError("could not sign token")
		c.JSON(jsonErr.StatusCode, jsonErr)
		c.Abort()
		return
	}

	user, err := repo.GetUserByEmail(claims.Email)
	if err != nil {
		jsonErr := models.NewNotFoundError(fmt.Sprintf("user with email %s not found", claims.Email))
		c.JSON(jsonErr.StatusCode, jsonErr)
		c.Abort()
		return
	}

	// save tokens to cache
	if err = cacheClient.Set(token.AccessUUID, user.ID, int(time.Unix(token.AccessTokenExpires, 0).Sub(time.Now()).Seconds())); err != nil {
		jsonErr := models.NewInternalServerError("could not update cache")
		c.JSON(jsonErr.StatusCode, jsonErr)
		c.Abort()
		return
	}

	if err = cacheClient.Set(token.RefreshUUID, user.ID, int(time.Unix(token.RefreshTokenExpires, 0).Sub(time.Now()).Seconds())); err != nil {
		jsonErr := models.NewInternalServerError("could not update cache")
		c.JSON(jsonErr.StatusCode, jsonErr)
		c.Abort()
		return
	}

	tokenResponse := &models.LoginResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}
	c.JSON(http.StatusOK, tokenResponse)
}
