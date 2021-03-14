package mailer

import "github.com/zbiljic/authzy/pkg/domain/user"

// Compile-time proof of interface implementation.
var _ Mailer = (*noopMailer)(nil)

type noopMailer struct{}

func (*noopMailer) ValidateEmail(email string) error {
	return nil
}

func (*noopMailer) Send(user *user.User, subject, body string, data map[string]interface{}) error {
	return nil
}

func (*noopMailer) ConfirmationMail(user *user.User, referrerURL string) error {
	return nil
}

func (*noopMailer) RecoveryMail(user *user.User, referrerURL string) error {
	return nil
}

func (*noopMailer) EmailChangeMail(user *user.User, referrerURL string) error {
	return nil
}
