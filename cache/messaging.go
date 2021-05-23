package cache

import (
	"github.com/go-redis/redis/v7"
	"log"
)

func HandleSignUpMessages(message *redis.Message) {
	if message == nil {
		return
	}

	log.Println(message.Payload)
}
