package store

import "github.com/bogdanrat/web-server/contracts/models"

type KeyValue interface {
	Get(key string) (interface{}, error)
	Put(*models.KeyValuePair) error
	Delete(key string) error
	GetAll() ([]*models.KeyValuePair, error)
}
