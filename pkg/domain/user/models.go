package user

import (
	"time"

	"github.com/zbiljic/authzy/pkg/jsonmap"
)

// User respresents a registered user.
type User struct {
	ID            string
	Email         string
	EmailVerified bool
	ValidSince    *time.Time

	Password          string
	PasswordHash      string
	PasswordUpdatedAt *time.Time

	Username           string
	NormalizedUsername string
	GivenName          string
	FamilyName         string
	Name               string
	Nickname           string
	Picture            string

	ConfirmationToken  string
	ConfirmationSentAt *time.Time

	RecoveryToken  string
	RecoverySentAt *time.Time

	EmailChangeToken  string
	EmailChange       string
	EmailChangeSentAt *time.Time

	AppMetaData  jsonmap.JSONMap
	UserMetaData jsonmap.JSONMap

	LastIP      string
	LastLoginAt *time.Time
	LoginsCount int64

	Blocked bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

// IsConfirmed checks if a user is already registered and confirmed.
func (u *User) IsConfirmed() bool {
	return u.EmailVerified
}
