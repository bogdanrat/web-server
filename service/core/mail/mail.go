package mail

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/bogdanrat/web-server/service/core/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	"time"
)

type Message struct {
	To      string
	Subject string
	Body    string
}

type Service struct {
	*gmail.Service
}

var (
	smtpService *Service
)

func NewService(smtpConfig config.SMTPConfig) error {
	oauth2Config := oauth2.Config{
		ClientID:     smtpConfig.ClientID,
		ClientSecret: smtpConfig.ClientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost",
	}

	token := oauth2.Token{
		AccessToken:  smtpConfig.AccessToken,
		RefreshToken: smtpConfig.RefreshToken,
		TokenType:    "Bearer",
		Expiry:       time.Now(),
	}

	var tokenSource = oauth2Config.TokenSource(context.Background(), &token)

	service, err := gmail.NewService(context.Background(), option.WithTokenSource(tokenSource))
	if err != nil {
		return fmt.Errorf("unable to create gmail service: %s", err)
	}

	smtpService = &Service{service}
	return nil
}

func Send(message *Message) error {
	emailTo := "To: " + message.To + "\r\n"
	subject := "Subject: " + message.Subject + "\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n"
	msg := []byte(emailTo + subject + mime + "\n" + message.Body)

	gMessage := &gmail.Message{
		Raw: base64.URLEncoding.EncodeToString(msg),
	}

	_, err := smtpService.Users.Messages.Send("me", gMessage).Do()
	if err != nil {
		return err
	}
	return nil
}
