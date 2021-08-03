package listener

import (
	"bytes"
	"fmt"
	"github.com/bogdanrat/web-server/contracts/models"
	"github.com/bogdanrat/web-server/service/core/i18n"
	"github.com/bogdanrat/web-server/service/core/mail"
	"github.com/bogdanrat/web-server/service/queue"
	"log"
)

type EventProcessor struct {
	EventListener queue.EventListener
	Translator    i18n.Translator
}

func NewEventProcessor(listener queue.EventListener, translator i18n.Translator) *EventProcessor {
	return &EventProcessor{
		EventListener: listener,
		Translator:    translator,
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
		Subject: p.Translator.Do(i18n.EmailWelcomeSubjectKey, nil),
	}

	emailSubstitutions := map[string]string{
		"username": user.Name,
	}
	buffer := bytes.Buffer{}
	buffer.WriteString(p.Translator.Do(i18n.EmailWelcomeBodyKey, emailSubstitutions))

	if event.QrImage != nil {
		buffer.WriteString(p.Translator.Do(i18n.EmailWelcomeBodyMFAKey, nil))
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
