package refreshtoken

import (
	"context"

	"github.com/zbiljic/authzy/pkg/domain/user"
)

type RefreshTokenUsecase interface {
	// GrantAuthenticatedUser creates a refresh token for the provided user.
	GrantAuthenticatedUser(context.Context, *user.User) (*RefreshToken, error)

	// GrantRefreshTokenSwap swaps a refresh token for a new one, revoking
	// the provided token.
	GrantRefreshTokenSwap(context.Context, *user.User, *RefreshToken) (*RefreshToken, error)

	// FindRefreshTokenByID retrieves an refresh token by ID.
	FindRefreshTokenByID(context.Context, string) (*RefreshToken, error)

	// FindRefreshTokenByToken retrieves an refresh token by token value.
	FindRefreshTokenByToken(context.Context, string) (*RefreshToken, error)

	// Revoke revokes the provided token.
	Revoke(context.Context, *RefreshToken) error

	// Logout deletes all refresh tokens for a user.
	Logout(context.Context, *user.User) error
}
