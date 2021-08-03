package store

import (
	"errors"
	"github.com/bogdanrat/web-server/contracts/models"
)

var (
	KeyNotFoundError = errors.New("key not found")
)

const (
	KeyIdentifier   = "Key"
	ValueIdentifier = "Value"
)

type KeyValue interface {
	Get(key string) (interface{}, error)
	Put(*models.KeyValuePair) error
	Delete(key string) error
	GetAll() ([]*models.KeyValuePair, error)
}
