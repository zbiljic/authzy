package mailer

import (
	_ "embed" // Remove this line to disable files embedding.
	"net/url"

	"github.com/netlify/mailme"

	"github.com/zbiljic/authzy/pkg/config"
	"github.com/zbiljic/authzy/pkg/domain/user"
	"github.com/zbiljic/authzy/pkg/logger"
)

type templateMailer struct {
	validateMailer

	Config *config.Config
	Mailer *mailme.Mailer
}

//go:embed templates/defaultConfirmationMail.go.html
var defaultConfirmationMail string

//go:embed templates/defaultRecoveryMail.go.html
var defaultRecoveryMail string

//go:embed templates/defaultEmailChangeMail.go.html
var defaultEmailChangeMail string

func newTemplateMailer(log logger.Logger, config *config.Config) Mailer {
	return &templateMailer{
		validateMailer: validateMailer{config: config.API.Mailer},
		Config:         config,
		Mailer: &mailme.Mailer{
			From:    config.SMTP.AdminEmail,
			Host:    config.SMTP.Host,
			Port:    config.SMTP.Port,
			User:    config.SMTP.User,
			Pass:    config.SMTP.Pass,
			BaseURL: config.SiteURL,
			Logger:  newLogrusLogger(log),
		},
	}
}

func (m *templateMailer) Send(user *user.User, subject, body string, data map[string]interface{}) error {
	return m.Mailer.Mail(
		user.Email,
		subject,
		"",
		body,
		data,
	)
}

func (m *templateMailer) ConfirmationMail(user *user.User, referrerURL string) error {
	query := url.Values{}
	query.Add("type", "signup")
	query.Add("token", user.ConfirmationToken)
	if len(referrerURL) > 0 {
		query.Add("redirect_to", referrerURL)
	}

	url, err := getSiteURL(referrerURL, m.Config.API.ExternalURL, m.Config.API.Mailer.URLPaths.Confirmation, query.Encode())
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"SiteURL":         m.Config.SiteURL,
		"ConfirmationURL": url,
		"Email":           user.Email,
		"Token":           user.ConfirmationToken,
		"Data":            user.UserMetaData,
	}

	return m.Mailer.Mail(
		user.Email,
		withDefault(m.Config.API.Mailer.Subjects.Confirmation, "Confirm Your Signup"),
		m.Config.API.Mailer.Templates.Confirmation,
		defaultConfirmationMail,
		data,
	)
}

func (m *templateMailer) RecoveryMail(user *user.User, referrerURL string) error {
	query := url.Values{}
	query.Add("type", "recovery")
	query.Add("token", user.RecoveryToken)
	if len(referrerURL) > 0 {
		query.Add("redirect_to", referrerURL)
	}

	url, err := getSiteURL(referrerURL, m.Config.API.ExternalURL, m.Config.API.Mailer.URLPaths.Recovery, query.Encode())
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"SiteURL":         m.Config.SiteURL,
		"ConfirmationURL": url,
		"Email":           user.Email,
		"Token":           user.RecoveryToken,
		"Data":            user.UserMetaData,
	}

	return m.Mailer.Mail(
		user.Email,
		withDefault(m.Config.API.Mailer.Subjects.Recovery, "Reset Your Password"),
		m.Config.API.Mailer.Templates.Recovery,
		defaultRecoveryMail,
		data,
	)
}

func (m *templateMailer) EmailChangeMail(user *user.User, referrerURL string) error {
	query := url.Values{}
	query.Add("type", "email_change")
	query.Add("email_change_token", user.RecoveryToken)
	if len(referrerURL) > 0 {
		query.Add("redirect_to", referrerURL)
	}

	url, err := getSiteURL(referrerURL, m.Config.API.ExternalURL, m.Config.API.Mailer.URLPaths.EmailChange, query.Encode())
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"SiteURL":         m.Config.SiteURL,
		"ConfirmationURL": url,
		"Email":           user.Email,
		"NewEmail":        user.EmailChange,
		"Token":           user.EmailChangeToken,
		"Data":            user.UserMetaData,
	}

	return m.Mailer.Mail(
		user.Email,
		withDefault(m.Config.API.Mailer.Subjects.EmailChange, "Confirm Email Change"),
		m.Config.API.Mailer.Templates.EmailChange,
		defaultEmailChangeMail,
		data,
	)
}
