package store

import (
	pb "github.com/bogdanrat/web-server/contracts/proto/storage_service"
	"io"
)

type Store interface {
	Init() error
	Put(key string, body io.Reader) error
	Get(key string, writer io.Writer) error
	GetAll() ([]*pb.StorageObject, error)
	Delete(fileName string) error
	DeleteAll(prefix ...string) error
}
