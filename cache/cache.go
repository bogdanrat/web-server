package cache

type Client interface {
	Set(key string, value interface{}, timeoutSeconds int) error
	Get(key string) (interface{}, error)
	Delete(key string) error
}

var (
	RedisClient Client
)
