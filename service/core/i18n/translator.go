package i18n

const (
	EmailWelcomeSubjectKey = "EMAIL_WELCOME_SUBJECT"
	EmailWelcomeBodyKey    = "EMAIL_WELCOME_BODY"
	EmailWelcomeBodyMFAKey = "EMAIL_WELCOME_BODY_MFA"
)

type Translator interface {
	Do(key string, substitutions map[string]string) string
	Reload() error
}
