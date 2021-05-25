package cache

import (
	"github.com/go-redis/redis/v7"
	"log"
)

func HandleAuthServiceMessages(message *redis.Message) {
	if message == nil {
		return
	}

	log.Println(message.Payload)
}
