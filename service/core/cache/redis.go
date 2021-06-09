package cache

import (
	"fmt"
	"github.com/bogdanrat/web-server/service/core/config"
	"github.com/go-redis/redis/v7"
	"log"
	"os"
	"time"
)

type Redis struct {
	redis         *redis.Client
	subscriptions map[string]*RedisSubscription
}

type RedisSubscription struct {
	subscription *redis.PubSub
	channels     []string
	closeChan    chan bool
}

func NewRedis(config config.RedisConfig) (*Redis, error) {
	if RedisClient != nil {
		return RedisClient.(*Redis), nil
	}

	host := os.Getenv("REDIS_HOST")
	if host == "" {
		log.Printf("env redis host not found, using config host %s\n", config.Host)
		host = config.Host
	}
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		log.Printf("env redis port not found, using config port %s\n", config.Port)
		port = config.Port
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: config.Password,
		DB:       0,
	})

	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}

	c := &Redis{
		redis:         client,
		subscriptions: make(map[string]*RedisSubscription),
	}

	RedisClient = c
	return RedisClient.(*Redis), nil
}

func (c *Redis) Set(key string, value interface{}, timeoutSeconds int) error {
	_, err := c.redis.Set(key, value, time.Duration(timeoutSeconds)*time.Second).Result()
	if err != nil {
		log.Printf("error writing to redis: %s", err)
	}
	return err
}

func (c *Redis) Get(key string) (interface{}, error) {
	result, err := c.redis.Get(key).Result()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Redis) Delete(key string) error {
	_, err := c.redis.Del(key).Result()
	return err
}

func (c *Redis) Publish(channel string, message interface{}) {
	c.redis.Publish(channel, message)
}

func (c *Redis) Subscribe(key string, handler func(message *redis.Message), channels ...string) *redis.PubSub {
	channels = append(channels, key)
	subscription := c.redis.Subscribe(channels...)

	redisSubscription := &RedisSubscription{
		subscription: subscription,
		channels:     channels,
		closeChan:    make(chan bool),
	}

	c.subscriptions[key] = redisSubscription
	go c.redisClientListener(redisSubscription, handler)

	return subscription
}

func (c *Redis) redisClientListener(redisSubscription *RedisSubscription, handler func(message *redis.Message)) {
	for {
		select {
		case <-redisSubscription.closeChan:
			return
		case message := <-redisSubscription.subscription.Channel():
			handler(message)
		}
	}
}

func (c *Redis) Unsubscribe(key string) {
	redisSubscription := c.subscriptions[key]
	if redisSubscription == nil {
		return
	}

	close(redisSubscription.closeChan)
	delete(c.subscriptions, key)
	_ = redisSubscription.subscription.Unsubscribe(redisSubscription.channels...)
}
