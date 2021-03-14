package mailer

import (
	"fmt"

	"github.com/badoux/checkmail"

	"github.com/zbiljic/authzy/pkg/config"
)

type validateMailer struct {
	noopMailer

	config *config.MailerConfig
}

func (m *validateMailer) ValidateEmail(email string) error {
	err := checkmail.ValidateFormat(email)
	if err != nil {
		return fmt.Errorf("validate email: %w", err)
	}

	if m.config.ValidateHost {
		err = checkmail.ValidateHost(email)
		if err != nil {
			return fmt.Errorf("validate email: %w", err)
		}
	}

	return nil
}
