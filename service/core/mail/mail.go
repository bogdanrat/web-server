package mail

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/bogdanrat/web-server/service/core/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	"net/http"
	"os"
	"time"
)

type Message struct {
	To         string
	Subject    string
	Body       string
	Attachment *Attachment
}

type Attachment struct {
	Name string
	Data []byte
}

type Service struct {
	*gmail.Service
}

var (
	smtpService *Service
)

func NewService(smtpConfig config.SMTPConfig) error {
	smtpConfig = sanitizeConfig(smtpConfig)

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

func sanitizeConfig(smtpConfig config.SMTPConfig) config.SMTPConfig {
	if smtpConfig.ClientID == "" {
		smtpConfig.ClientID = os.Getenv("SMTP_CLIENT_ID")
	}
	if smtpConfig.ClientSecret == "" {
		smtpConfig.ClientSecret = os.Getenv("SMTP_CLIENT_SECRET")
	}
	if smtpConfig.AccessToken == "" {
		smtpConfig.AccessToken = os.Getenv("SMTP_ACCESS_TOKEN")
	}
	if smtpConfig.RefreshToken == "" {
		smtpConfig.RefreshToken = os.Getenv("SMTP_REFRESH_TOKEN")
	}

	return smtpConfig
}

func Send(message *Message) error {
	hasAttachment := false
	if message.Attachment != nil {
		hasAttachment = true
	}

	boundary := randStr(32, "alphanum")

	messageBody := bytes.Buffer{}
	messageBody.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s \n", boundary))
	messageBody.WriteString("MIME-Version: 1.0\n")
	messageBody.WriteString(fmt.Sprintf("To: %s\n", message.To))
	messageBody.WriteString(fmt.Sprintf("Subject: %s \n\n", message.Subject))
	if hasAttachment {
		messageBody.WriteString(fmt.Sprintf("--%s\n", boundary))
	}
	messageBody.WriteString("Content-Type: text/html; charset=\"UTF-8\"\n")
	messageBody.WriteString("MIME-Version: 1.0\n")
	messageBody.WriteString("Content-Transfer-Encoding: 7bit\n\n")
	messageBody.WriteString(fmt.Sprintf("%s\n\n", message.Body))
	if hasAttachment {
		messageBody.WriteString(fmt.Sprintf("--%s\n", boundary))
	}
	if hasAttachment {
		fileMIMEType := http.DetectContentType(message.Attachment.Data)
		fileData := base64.StdEncoding.EncodeToString(message.Attachment.Data)
		messageBody.WriteString(fmt.Sprintf("Content-Type: %s"+"; name=\"%s\"\n", fileMIMEType, message.Attachment.Name))
		messageBody.WriteString("MIME-Version: 1.0\n")
		messageBody.WriteString("Content-Transfer-Encoding: base64\n")
		messageBody.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\n\n", message.Attachment.Name))
		messageBody.WriteString(chunkSplit(fileData, 76, "\n"))
		messageBody.WriteString(fmt.Sprintf("--%s--", boundary))
	}

	gMessage := &gmail.Message{
		Raw: base64.URLEncoding.EncodeToString(messageBody.Bytes()),
	}

	_, err := smtpService.Users.Messages.Send("me", gMessage).Do()
	if err != nil {
		return err
	}
	return nil
}

func randStr(strSize int, randType string) string {
	var dictionary string

	if randType == "alphanum" {
		dictionary = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	}

	var strBytes = make([]byte, strSize)
	_, _ = rand.Read(strBytes)

	for k, v := range strBytes {
		strBytes[k] = dictionary[v%byte(len(dictionary))]
	}

	return string(strBytes)
}

func chunkSplit(body string, limit int, end string) string {
	var charSlice []rune

	// push characters to slice
	for _, char := range body {
		charSlice = append(charSlice, char)
	}

	var result = ""

	for len(charSlice) >= 1 {
		// convert slice/array back to string
		// but insert end at specified limit
		result = result + string(charSlice[:limit]) + end

		// discard the elements that were copied over to result
		charSlice = charSlice[limit:]

		// change the limit
		// to cater for the last few words in
		if len(charSlice) < limit {
			limit = len(charSlice)
		}
	}
	return result
}
