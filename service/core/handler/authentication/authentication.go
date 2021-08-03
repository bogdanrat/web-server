package authentication

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bogdanrat/web-server/contracts/models"
	authPb "github.com/bogdanrat/web-server/contracts/proto/auth_service"
	"github.com/bogdanrat/web-server/service/core/cache"
	"github.com/bogdanrat/web-server/service/core/config"
	"github.com/bogdanrat/web-server/service/core/forms"
	"github.com/bogdanrat/web-server/service/core/lib"
	"github.com/bogdanrat/web-server/service/core/render"
	"github.com/bogdanrat/web-server/service/core/store"
	"github.com/bogdanrat/web-server/service/core/util"
	"github.com/bogdanrat/web-server/service/queue"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"net/http"
	"reflect"
	"strconv"
	"time"
)

type RPCConfig struct {
	Client      authPb.AuthClient
	CallOptions []grpc.CallOption
	Deadline    int64
}

type Handler struct {
	Repository   store.DatabaseRepository
	Cache        cache.Client
	AuthService  *RPCConfig
	EventEmitter queue.EventEmitter
}

func NewHandler(repo store.DatabaseRepository, cacheClient cache.Client, authConfig *RPCConfig, eventEmitter queue.EventEmitter) *Handler {
	return &Handler{
		Repository:   repo,
		Cache:        cacheClient,
		AuthService:  authConfig,
		EventEmitter: eventEmitter,
	}
}

func (h *Handler) ShowLogin(c *gin.Context) {
	templateData := &models.TemplateData{}
	if config.AppConfig.Authentication.MFA {
		intMap := make(map[string]int)
		intMap["mfa"] = 1
		templateData.IntMap = intMap
	}
	_ = render.Template(c.Writer, c.Request, "login.page.tmpl", templateData)
}

func (h *Handler) SignUp(c *gin.Context) {
	request := &models.SignUpRequest{}
	if err := c.ShouldBindJSON(request); err != nil {
		jsonErr := models.NewBadRequestError("invalid sign up request")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	if _, err := h.Repository.GetUserByEmail(request.Email); err == nil {
		jsonErr := models.NewBadRequestError(fmt.Sprintf("email %s already registered", request.Email), "email")
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

	var qrImage []byte

	// If MFA is enabled, send a request to get a QR code
	if config.AppConfig.Authentication.MFA {
		// Deadlines: the entire request chain needs to respond by the deadline set by the app that initiated the request.
		// Timeouts: applied at each RPC, at each service invocation, not for the entire life cycle of the request.
		deadline := time.Now().Add(time.Millisecond * time.Duration(h.AuthService.Deadline))
		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		defer cancel()

		response, err := h.AuthService.Client.GenerateQRCode(
			ctx,
			&authPb.GenerateQRCodeRequest{Email: request.Email},
			h.AuthService.CallOptions...,
		)
		if jsonErr := lib.HandleRPCError(err); jsonErr != nil {
			c.JSON(jsonErr.StatusCode, jsonErr)
			return
		}

		user.QRSecret = &response.Secret
		qrImage = response.Image

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
		userJson, err := json.Marshal(user)
		if err != nil {
			jsonErr := models.NewInternalServerError(fmt.Sprintf("unable to json marshal user struct: %s", err))
			c.JSON(jsonErr.StatusCode, jsonErr)
			return
		}
		h.Cache.(*cache.Redis).Publish(config.AppConfig.Authentication.Channel, userJson)
	}

	err := h.EventEmitter.Emit(&models.UserSignUpEvent{
		User:    user,
		QrImage: qrImage,
	})
	if err != nil {
		jsonErr := models.NewInternalServerError(fmt.Sprintf("cannot emit user sign up event: %s", err))
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}
}

func (h *Handler) Login(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		jsonErr := models.NewInternalServerError("could not parse form")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	email := c.Request.Form.Get("email")
	password := c.Request.Form.Get("password")
	qrCode := c.Request.Form.Get("qr_code")

	form := forms.New(c.Request.PostForm)
	requiredFields := []string{"email", "password"}
	if config.AppConfig.Authentication.MFA {
		requiredFields = append(requiredFields, "qr_code")
	}
	form.Required(requiredFields...)
	form.ValidEmail("email")

	if !form.Valid() {
		var jsonErr *models.JSONError
		formJson, err := form.Marshal()
		if err == nil {
			jsonErr = models.NewBadRequestError(string(formJson))
		} else {
			jsonErr = models.NewBadRequestError("invalid form submitted")
		}
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	user, err := h.Repository.GetUserByEmail(email)
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

		deadline := time.Now().Add(time.Millisecond * time.Duration(h.AuthService.Deadline))
		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		defer cancel()

		response, err := h.AuthService.Client.ValidateQRCode(
			ctx,
			&authPb.ValidateQRCodeRequest{QrCode: qrCode, QrSecret: *user.QRSecret},
			h.AuthService.CallOptions...,
		)
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

	err = user.CheckPassword(password)
	if err != nil {
		jsonErr := models.NewUnauthorizedError("invalid user credentials", "password")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	deadline := time.Now().Add(time.Millisecond * time.Duration(h.AuthService.Deadline))
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	response, err := h.AuthService.Client.GenerateToken(
		ctx,
		&authPb.GenerateTokenRequest{
			Email:                email,
			AccessTokenDuration:  config.AppConfig.Authentication.AccessTokenDuration,
			RefreshTokenDuration: config.AppConfig.Authentication.RefreshTokenDuration,
		},
		h.AuthService.CallOptions...,
	)
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

	deadline := time.Now().Add(time.Millisecond * time.Duration(h.AuthService.Deadline))
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	response, err := h.AuthService.Client.ValidateAccessToken(
		ctx,
		&authPb.ValidateAccessTokenRequest{SignedToken: token},
		h.AuthService.CallOptions...,
	)
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

	deadline := time.Now().Add(time.Millisecond * time.Duration(h.AuthService.Deadline))
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	validateResponse, err := h.AuthService.Client.ValidateRefreshToken(
		ctx,
		&authPb.ValidateRefreshTokenRequest{SignedToken: refreshToken},
		h.AuthService.CallOptions...,
	)
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

	ctx, cancel = context.WithDeadline(context.Background(), deadline)
	defer cancel()

	// create new pairs of access & refresh tokens
	generateResponse, err := h.AuthService.Client.GenerateToken(
		ctx,
		&authPb.GenerateTokenRequest{
			Email:                validateResponse.Email,
			AccessTokenDuration:  2,
			RefreshTokenDuration: 1140,
		},
		h.AuthService.CallOptions...,
	)
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
