package store

import (
	"errors"
	"github.com/bogdanrat/web-server/contracts/models"
	"github.com/bogdanrat/web-server/service/core/store"
	"github.com/bogdanrat/web-server/service/queue"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type Handler struct {
	Store        store.KeyValue
	EventEmitter queue.EventEmitter
}

func NewHandler(store store.KeyValue, eventEmitter queue.EventEmitter) *Handler {
	return &Handler{
		Store:        store,
		EventEmitter: eventEmitter,
	}
}

func (h *Handler) GetPair(c *gin.Context) {
	request := &models.GetPairRequest{}
	if err := c.ShouldBind(request); err != nil {
		jsonErr := models.NewBadRequestError("key is required", "key")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	value, err := h.Store.Get(request.Key)
	if err != nil {
		var jsonErr *models.JSONError
		if errors.Is(err, store.KeyNotFoundError) {
			jsonErr = models.NewNotFoundError("key not found", "key")

		} else {
			jsonErr = models.NewInternalServerError(err.Error())
		}
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	stringValue, ok := value.(string)
	if !ok {
		jsonErr := models.NewInternalServerError("could not convert value to string")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	response := &models.GetPairResponse{
		Key:   request.Key,
		Value: stringValue,
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetPairs(c *gin.Context) {
	pairs, err := h.Store.GetAll()
	if err != nil {
		jsonErr := models.NewInternalServerError(err.Error())
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}
	c.JSON(http.StatusOK, pairs)
}

func (h *Handler) PostPairs(c *gin.Context) {
	request := make([]*models.KeyValuePair, 0)
	if err := c.ShouldBind(&request); err != nil {
		jsonErr := models.NewBadRequestError("key and value are required", "key", "value")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	if err := h.Store.PutMany(request); err != nil {
		jsonErr := models.NewInternalServerError(err.Error())
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	if err := h.EventEmitter.Emit(&models.NewKeyValuePairEvent{
		Pairs: request,
	}); err != nil {
		log.Printf("cannot emit new key value par event: %s", err)
	}

	c.Status(http.StatusCreated)
}

func (h *Handler) DeletePair(c *gin.Context) {
	request := &models.DeletePairRequest{}
	if err := c.ShouldBind(request); err != nil {
		jsonErr := models.NewBadRequestError("key is required", "key")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	err := h.Store.Delete(request.Key)
	if err != nil {
		jsonErr := models.NewInternalServerError(err.Error())
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	if err := h.EventEmitter.Emit(&models.DeleteKeyValuePairEvent{
		Key: request.Key,
	}); err != nil {
		log.Printf("cannot emit new key value par event: %s", err)
	}

	c.Status(http.StatusOK)
}
