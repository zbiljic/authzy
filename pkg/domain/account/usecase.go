package account

import "context"

type AccountUsecase interface {
	// CreateAccount creates new account in the system.
	CreateAccount(context.Context, *Account) (*Account, error)

	// UpdateAccount updates existing account.
	UpdateAccount(context.Context, *Account) (*Account, error)

	// FindAllForUser retrieves all accounts for specified user ID.
	FindAllForUser(ctx context.Context, userID string) ([]*Account, error)
}
