package api

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/zbiljic/authzy/pkg/domain/user"
	"github.com/zbiljic/authzy/pkg/mailer"
	"github.com/zbiljic/authzy/pkg/ulid"
)

var (
	ErrMaxFrequencyLimit error = errors.New("Frequency limit reached")
)

func (s *server) validateEmail(ctx context.Context, email string) error {
	if email == "" {
		return unprocessableEntityError("An email address is required")
	}

	mailer := s.Mailer(ctx)
	if err := mailer.ValidateEmail(email); err != nil {
		return unprocessableEntityError("Unable to validate email address: " + err.Error())
	}

	return nil
}

func (s *server) sendConfirmation(ctx context.Context, u *user.User, mailer mailer.Mailer, maxFrequency time.Duration, referrerURL string) error {
	now := time.Now()

	if u.ConfirmationSentAt != nil && !u.ConfirmationSentAt.Add(maxFrequency).Before(now) {
		return ErrMaxFrequencyLimit
	}

	oldToken := u.ConfirmationToken

	u.ConfirmationToken = ulid.ULID().String()

	if err := mailer.ConfirmationMail(u, referrerURL); err != nil {
		u.ConfirmationToken = oldToken

		return fmt.Errorf("error sending confirmation email: %w", err)
	}

	u.ConfirmationSentAt = &now

	_, err := s.userUsecase.UpdateUser(ctx, u)
	if err != nil {
		return fmt.Errorf("database error updating user for confirmation: %w", err)
	}

	return nil
}

func (s *server) sendPasswordRecovery(ctx context.Context, u *user.User, mailer mailer.Mailer, maxFrequency time.Duration, referrerURL string) error {
	now := time.Now()

	if u.RecoverySentAt != nil && !u.RecoverySentAt.Add(maxFrequency).Before(now) {
		return ErrMaxFrequencyLimit
	}

	oldToken := u.RecoveryToken

	u.RecoveryToken = ulid.ULID().String()

	if err := mailer.RecoveryMail(u, referrerURL); err != nil {
		u.RecoveryToken = oldToken

		return fmt.Errorf("error sending recovery email: %w", err)
	}

	u.RecoverySentAt = &now

	_, err := s.userUsecase.UpdateUser(ctx, u)
	if err != nil {
		return fmt.Errorf("database error updating user for recovery: %w", err)
	}

	return nil
}

func (s *server) sendEmailChange(ctx context.Context, u *user.User, mailer mailer.Mailer, email string, referrerURL string) error {
	now := time.Now()

	oldToken := u.EmailChangeToken
	oldEmail := u.Email

	u.EmailChangeToken = ulid.ULID().String()
	u.EmailChange = email

	if err := mailer.EmailChangeMail(u, referrerURL); err != nil {
		u.EmailChangeToken = oldToken
		u.EmailChange = oldEmail

		return fmt.Errorf("error sending email change email: %w", err)
	}

	u.EmailChangeSentAt = &now

	_, err := s.userUsecase.UpdateUser(ctx, u)
	if err != nil {
		return fmt.Errorf("database error updating user for email change: %w", err)
	}

	return nil
}
