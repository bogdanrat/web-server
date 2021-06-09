package listener

import (
	"bytes"
	"fmt"
	"github.com/bogdanrat/web-server/contracts/models"
	"github.com/bogdanrat/web-server/service/core/mail"
	"github.com/bogdanrat/web-server/service/queue"
	"log"
)

type EventProcessor struct {
	EventListener queue.EventListener
}

func NewEventProcessor(listener queue.EventListener) *EventProcessor {
	return &EventProcessor{
		EventListener: listener,
	}
}

func (p *EventProcessor) ProcessEvent() error {
	received, errors, err := p.EventListener.Listen(models.UserSignUpEventName)
	if err != nil {
		return fmt.Errorf("could not listen for events: %s", err)
	}

	for {
		select {
		case event := <-received:
			p.handleEvent(event)
		case err := <-errors:
			log.Printf("received error while processing message: %s", err)
		}
	}
}

func (p *EventProcessor) handleEvent(event queue.Event) {
	switch e := event.(type) {
	case *models.UserSignUpEvent:
		p.handleUserSignUpEvent(e)
	default:
		log.Printf("unknown event: %t", e)
	}
}

func (p *EventProcessor) handleUserSignUpEvent(event *models.UserSignUpEvent) {
	user := event.User
	if user == nil {
		log.Printf("event user field is nil")
		return
	}

	email := &mail.Message{
		To:      user.Email,
		Subject: "Welcome",
	}

	buffer := bytes.Buffer{}
	buffer.WriteString(fmt.Sprintf("Welcome to Web Server App %s!\n", user.Name))

	if event.QrImage != nil {
		buffer.WriteString("Please scan the attached QR Code in Google Authenticator and use the generated codes to login in.\n")
		email.Attachment = &mail.Attachment{
			Name: "Personal QR Code",
			Data: event.QrImage,
		}
	}

	email.Body = buffer.String()

	if err := mail.Send(email); err != nil {
		log.Printf("cannot send email: %s", err)
	}

	log.Printf("Sent Welcome Email to %s\n", user.Email)
}
