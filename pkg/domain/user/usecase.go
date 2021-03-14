package user

import (
	"context"
	"net"
)

type UserUsecase interface {
	// CreateUser creates new user in the system.
	CreateUser(context.Context, *User) (*User, error)

	// UpdateUser updates existing user.
	UpdateUser(context.Context, *User) (*User, error)

	// UpdatePassword updates existing user password.
	UpdatePassword(ctx context.Context, id string, password []byte) (*User, error)

	// ConfirmUser confirms existing user.
	ConfirmUser(ctx context.Context, id string) (*User, error)

	// ConfirmRecovery confirms recovery of user credentials via email.
	ConfirmRecovery(context.Context, *User) (*User, error)

	// ConfirmEmailChange confirms the change of email for a user.
	ConfirmEmailChange(context.Context, *User) (*User, error)

	// FindUserByID retrieves an user by ID.
	FindUserByID(context.Context, string) (*User, error)

	// FindUserByEmail retrieves an user by email.
	FindUserByEmail(context.Context, string) (*User, error)

	// FindUserByConfirmationToken finds user with the matching confirmation token.
	FindUserByConfirmationToken(context.Context, string) (*User, error)

	// FindUserByRecoveryToken finds a user with the matching recovery token.
	FindUserByRecoveryToken(context.Context, string) (*User, error)

	// Authenticate a user using password.
	Authenticate(ctx context.Context, identifier string, password []byte) (*User, error)

	// UserSignedIn updates last sign in time.
	UserSignedIn(ctx context.Context, user *User, ipAddress net.IP) (*User, error)

	// UpdateUserMetaData sets all user data from a map of updates, ensuring
	// that it doesn't override attributes that are not in the provided map.
	UpdateUserMetaData(ctx context.Context, user *User, updates map[string]interface{}) (*User, error)

	// UpdateAppMetaData updates all app data from a map of updates.
	UpdateAppMetaData(ctx context.Context, user *User, updates map[string]interface{}) (*User, error)
}
