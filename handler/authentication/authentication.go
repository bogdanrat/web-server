package authentication

import (
	"context"
	"fmt"
	"github.com/bogdanrat/web-server/cache"
	"github.com/bogdanrat/web-server/config"
	"github.com/bogdanrat/web-server/lib"
	"github.com/bogdanrat/web-server/models"
	"github.com/bogdanrat/web-server/repository"
	pb "github.com/bogdanrat/web-server/service/auth/proto"
	"github.com/bogdanrat/web-server/util"
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
	"strconv"
	"time"
)

type Handler struct {
	Repository repository.DatabaseRepository
	Cache      cache.Client
	RPC        pb.AuthClient
}

func NewHandler(repo repository.DatabaseRepository, cacheClient cache.Client, authClient pb.AuthClient) *Handler {
	return &Handler{
		Repository: repo,
		Cache:      cacheClient,
		RPC:        authClient,
	}
}

func (h *Handler) SignUp(c *gin.Context) {
	request := &models.SignUpRequest{}
	if err := c.ShouldBindJSON(request); err != nil {
		jsonErr := models.NewBadRequestError("invalid sign up request")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	if _, err := h.Repository.GetUserByEmail(request.Email); err == nil {
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

	// If MFA is enabled, send a request to get a QR code
	if config.AppConfig.Authentication.MFA {
		// Deadlines: the entire request chain needs to respond by the deadline set by the app that initiated the request.
		// Timeouts: applied at each RPC, at each service invocation, not for the entire life cycle of the request.
		deadline := time.Now().Add(time.Second * 2)
		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		defer cancel()

		response, err := h.RPC.GenerateQRCode(ctx, &pb.GenerateQRCodeRequest{Email: request.Email})
		if jsonErr := lib.HandleRPCError(err); jsonErr != nil {
			c.JSON(jsonErr.StatusCode, jsonErr)
			return
		}

		user.QRSecret = &response.Secret
		qrImage := response.Image

		c.Writer.Header().Set("Content-Type", "image/png")
		c.Writer.Header().Set("Content-Length", strconv.Itoa(len(qrImage)))
		if _, err = c.Writer.Write(qrImage); err != nil {
			c.JSON(http.StatusInternalServerError, "unable to write image")
			return
		}
	}

	if err := h.Repository.InsertUser(user); err != nil {
		jsonErr := models.NewInternalServerError("unable to insert user into db")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	if reflect.ValueOf(h.Cache).Elem().Type().AssignableTo(reflect.TypeOf(cache.Redis{})) {
		h.Cache.(*cache.Redis).Publish(config.AppConfig.Authentication.Channel, user.Email)
	}
}

func (h *Handler) Login(c *gin.Context) {
	request := &models.LoginRequest{}
	if err := c.ShouldBindJSON(request); err != nil {
		jsonErr := models.NewBadRequestError("invalid login request")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	user, err := h.Repository.GetUserByEmail(request.Email)
	if err != nil {
		jsonErr := models.NewNotFoundError("user not found", "email")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	if config.AppConfig.Authentication.MFA {
		if user.QRSecret == nil {
			jsonErr := models.NewUnprocessableEntityError("missing qr code when mfa is enabled", "qr_secret", "")
			c.JSON(jsonErr.StatusCode, jsonErr)
			return
		}

		deadline := time.Now().Add(time.Second * 2)
		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		defer cancel()

		response, err := h.RPC.ValidateQRCode(ctx, &pb.ValidateQRCodeRequest{QrCode: request.QRCode, QrSecret: *user.QRSecret})
		if jsonErr := lib.HandleRPCError(err); jsonErr != nil {
			c.JSON(jsonErr.StatusCode, jsonErr)
			return
		}

		if !response.Validated {
			jsonErr := models.NewUnauthorizedError("invalid qr code", "qr_code")
			c.JSON(jsonErr.StatusCode, jsonErr)
			return
		}
	}

	err = user.CheckPassword(request.Password)
	if err != nil {
		jsonErr := models.NewUnauthorizedError("invalid user credentials", "password")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	response, err := h.RPC.GenerateToken(context.Background(), &pb.GenerateTokenRequest{
		Email: request.Email,
		// TODO: take from config
		AccessTokenDuration:  2,
		RefreshTokenDuration: 1140,
	})
	if jsonErr := lib.HandleRPCError(err); jsonErr != nil {
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	token := response.Token

	if err = h.Cache.Set(token.AccessUuid, user.ID, int(time.Unix(token.AccessTokenExpires, 0).Sub(time.Now()).Seconds())); err != nil {
		jsonErr := models.NewInternalServerError("could not update cache")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	if err = h.Cache.Set(token.RefreshUuid, user.ID, int(time.Unix(token.RefreshTokenExpires, 0).Sub(time.Now()).Seconds())); err != nil {
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

func (h *Handler) Logout(c *gin.Context) {
	token, jsonErr := util.ExtractToken(c.Request)
	if jsonErr != nil {
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	response, err := h.RPC.ValidateAccessToken(context.Background(), &pb.ValidateAccessTokenRequest{SignedToken: token})
	if jsonErr = lib.HandleRPCError(err); jsonErr != nil {
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	_, err = h.Cache.Get(response.AccessUuid)
	if err != nil {
		jsonErr = models.NewAlreadyReportedError("not logged in")
		c.JSON(jsonErr.StatusCode, jsonErr)
		c.Abort()
		return
	}

	err = h.Cache.Delete(response.AccessUuid)
	if err != nil {
		jsonErr = models.NewInternalServerError("could not delete access token from cache")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	c.JSON(http.StatusOK, "Successfully logged out")
}

func (h *Handler) RefreshToken(c *gin.Context) {
	mapToken := make(map[string]string, 0)
	if err := c.ShouldBindJSON(&mapToken); err != nil {
		jsonErr := models.NewBadRequestError("token malformed", "refresh_token")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	refreshToken := mapToken["refresh_token"]
	validateResponse, err := h.RPC.ValidateRefreshToken(context.Background(), &pb.ValidateRefreshTokenRequest{SignedToken: refreshToken})
	if jsonErr := lib.HandleRPCError(err); jsonErr != nil {
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	// delete the previous refresh token
	err = h.Cache.Delete(validateResponse.RefreshUuid)
	if err != nil {
		jsonErr := models.NewInternalServerError("could not delete refresh token from cache")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	// create new pairs of access & refresh tokens
	generateResponse, err := h.RPC.GenerateToken(context.Background(), &pb.GenerateTokenRequest{
		Email:                validateResponse.Email,
		AccessTokenDuration:  2,
		RefreshTokenDuration: 1140,
	})
	if jsonErr := lib.HandleRPCError(err); jsonErr != nil {
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	token := generateResponse.Token

	user, err := h.Repository.GetUserByEmail(validateResponse.Email)
	if err != nil {
		jsonErr := models.NewNotFoundError(fmt.Sprintf("user with email %s not found", validateResponse.Email))
		c.JSON(jsonErr.StatusCode, jsonErr)
		c.Abort()
		return
	}

	// save tokens to cache
	if err = h.Cache.Set(token.AccessUuid, user.ID, int(time.Unix(token.AccessTokenExpires, 0).Sub(time.Now()).Seconds())); err != nil {
		jsonErr := models.NewInternalServerError("could not update cache")
		c.JSON(jsonErr.StatusCode, jsonErr)
		c.Abort()
		return
	}

	if err = h.Cache.Set(token.RefreshUuid, user.ID, int(time.Unix(token.RefreshTokenExpires, 0).Sub(time.Now()).Seconds())); err != nil {
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
