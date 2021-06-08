package cache

import (
	"encoding/json"
	"fmt"
	"github.com/bogdanrat/web-server/contracts/models"
	"github.com/bogdanrat/web-server/service/core/mail"
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

	email := &mail.Message{
		To:      user.Email,
		Subject: "Welcome",
		Body:    fmt.Sprintf("Welcome to Web Server App %s!", user.Name),
	}
	if err := mail.Send(email); err != nil {
		log.Printf("cannot send email: %s", err)
	}

	log.Printf("Sent Welcome Email to %s\n", user.Email)
}
