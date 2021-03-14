package mailer

import (
	"net/url"

	"github.com/zbiljic/authzy/pkg/config"
	"github.com/zbiljic/authzy/pkg/domain/user"
	"github.com/zbiljic/authzy/pkg/logger"
)

// Mailer defines the interface a mailer must implement.
type Mailer interface {
	// ValidateEmail returns nil if the email is valid, otherwise an error
	// indicating the reason it is invalid.
	ValidateEmail(email string) error

	// Send can be used to send one-off emails to users.
	Send(user *user.User, subject, body string, data map[string]interface{}) error

	// ConfirmationMail sends a signup confirmation mail to a new user.
	ConfirmationMail(user *user.User, referrerURL string) error

	// RecoveryMail sends a password recovery mail.
	RecoveryMail(user *user.User, referrerURL string) error

	// EmailChangeMail sends an email change confirmation mail to a user.
	EmailChangeMail(user *user.User, referrerURL string) error
}

// NewMailer returns a new mailer.
func NewMailer(log logger.Logger, config *config.Config) Mailer {
	if config.SMTP == nil || config.SMTP.Host == "" {
		return &validateMailer{config: config.API.Mailer}
	}

	return newTemplateMailer(log, config)
}

func withDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}

	return value
}

func getSiteURL(referrerURL, siteURL, filepath, rawQuery string) (string, error) {
	baseURL := siteURL
	if filepath == "" && referrerURL != "" {
		baseURL = referrerURL
	}

	site, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	if filepath != "" {
		path, err := url.Parse(filepath)
		if err != nil {
			return "", err
		}
		site = site.ResolveReference(path)
	}
	if rawQuery != "" {
		site.RawQuery = rawQuery
	}

	return site.String(), nil
}
