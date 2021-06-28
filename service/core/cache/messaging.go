package cache

import (
	"encoding/json"
	"github.com/bogdanrat/web-server/contracts/models"
	"github.com/go-redis/redis/v7"
	"log"
)

func HandleAuthServiceMessages(message *redis.Message) {
	if message == nil {
		return
	}

	user := &models.User{}
	err := json.Unmarshal([]byte(message.Payload), user)
	if err != nil {
		return
	}

	// Replaced with SQS
	log.Printf("Should send Welcome Email to %s\n", user.Email)
}
