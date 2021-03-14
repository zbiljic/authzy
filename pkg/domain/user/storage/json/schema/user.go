package schema

import (
	"time"

	"github.com/zbiljic/authzy/pkg/domain/user"
	"github.com/zbiljic/authzy/pkg/jsonmap"
)

type User struct {
	ID            string     `json:"user_id" validate:"required,alphanum"`
	Email         string     `json:"email,omitempty" validate:"required,lowercase,email"`
	EmailVerified bool       `json:"email_verified,omitempty"`
	ValidSince    *time.Time `json:"valid_since,omitempty"`

	PasswordHash      string     `json:"password_hash,omitempty"`
	PasswordUpdatedAt *time.Time `json:"password_updated_at,omitempty"`

	Username           string `json:"username,omitempty" validate:"required,username"`
	NormalizedUsername string `json:"normalized_username,omitempty" validate:"required,lowercase,username"`
	GivenName          string `json:"given_name,omitempty"`
	FamilyName         string `json:"family_name,omitempty"`
	Name               string `json:"name,omitempty"`
	Nickname           string `json:"nickname,omitempty"`
	Picture            string `json:"picture,omitempty" validate:"omitempty,url"`

	ConfirmationToken  string     `json:"confirmation_token,omitempty"`
	ConfirmationSentAt *time.Time `json:"confirmation_sent_at,omitempty"`

	RecoveryToken  string     `json:"recovery_token,omitempty"`
	RecoverySentAt *time.Time `json:"recovery_sent_at,omitempty"`

	EmailChangeToken  string     `json:"email_change_token,omitempty"`
	EmailChange       string     `json:"new_email,omitempty"`
	EmailChangeSentAt *time.Time `json:"email_change_sent_at,omitempty"`

	AppMetaData  jsonmap.JSONMap `json:"app_metadata,omitempty"`
	UserMetaData jsonmap.JSONMap `json:"user_metadata,omitempty"`

	LastIP      string     `json:"last_ip,omitempty"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	LoginsCount int64      `json:"logins_count,omitempty"`

	Blocked bool `json:"blocked,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) BeforeSave() error {
	if u.ValidSince != nil && u.ValidSince.IsZero() {
		u.ValidSince = nil
	}
	if u.PasswordUpdatedAt != nil && u.PasswordUpdatedAt.IsZero() {
		u.PasswordUpdatedAt = nil
	}
	if u.ConfirmationSentAt != nil && u.ConfirmationSentAt.IsZero() {
		u.ConfirmationSentAt = nil
	}
	if u.RecoverySentAt != nil && u.RecoverySentAt.IsZero() {
		u.RecoverySentAt = nil
	}
	if u.EmailChangeSentAt != nil && u.EmailChangeSentAt.IsZero() {
		u.EmailChangeSentAt = nil
	}
	if u.LastLoginAt != nil && u.LastLoginAt.IsZero() {
		u.LastLoginAt = nil
	}

	return nil
}

func UserToSchema(in *user.User) *User {
	out := &User{}
	if in != nil {
		out.ID = in.ID
		out.Email = in.Email
		out.EmailVerified = in.EmailVerified
		out.ValidSince = in.ValidSince
		out.PasswordHash = in.PasswordHash
		out.PasswordUpdatedAt = in.PasswordUpdatedAt
		out.Username = in.Username
		out.NormalizedUsername = in.NormalizedUsername
		out.Picture = in.Picture
		out.ConfirmationToken = in.ConfirmationToken
		out.ConfirmationSentAt = in.ConfirmationSentAt
		out.RecoveryToken = in.RecoveryToken
		out.RecoverySentAt = in.RecoverySentAt
		out.EmailChangeToken = in.EmailChangeToken
		out.EmailChange = in.EmailChange
		out.EmailChangeSentAt = in.EmailChangeSentAt
		out.AppMetaData = in.AppMetaData
		out.UserMetaData = in.UserMetaData
		out.LastIP = in.LastIP
		out.LastLoginAt = in.LastLoginAt
		out.LoginsCount = in.LoginsCount
		out.Blocked = in.Blocked
		out.CreatedAt = in.CreatedAt
		out.UpdatedAt = in.UpdatedAt
	}

	return out
}

func UserFromSchema(in *User) *user.User {
	out := &user.User{}
	out.ID = in.ID
	out.Email = in.Email
	out.EmailVerified = in.EmailVerified
	out.ValidSince = in.ValidSince
	out.PasswordHash = in.PasswordHash
	out.PasswordUpdatedAt = in.PasswordUpdatedAt
	out.Username = in.Username
	out.NormalizedUsername = in.NormalizedUsername
	out.Picture = in.Picture
	out.ConfirmationToken = in.ConfirmationToken
	out.ConfirmationSentAt = in.ConfirmationSentAt
	out.RecoveryToken = in.RecoveryToken
	out.RecoverySentAt = in.RecoverySentAt
	out.EmailChangeToken = in.EmailChangeToken
	out.EmailChange = in.EmailChange
	out.EmailChangeSentAt = in.EmailChangeSentAt
	out.AppMetaData = in.AppMetaData
	out.UserMetaData = in.UserMetaData
	out.LastIP = in.LastIP
	out.LastLoginAt = in.LastLoginAt
	out.LoginsCount = in.LoginsCount
	out.Blocked = in.Blocked
	out.CreatedAt = in.CreatedAt
	out.UpdatedAt = in.UpdatedAt

	return out
}
