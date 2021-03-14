package usecases

import (
	"context"

	"github.com/zbiljic/authzy/pkg/domain/refreshtoken"
	"github.com/zbiljic/authzy/pkg/domain/user"
)

// Compile-time proof of interface implementation.
var _ refreshtoken.RefreshTokenUsecase = (*noopRefreshTokenUsecase)(nil)

// noopRefreshTokenUsecase can be embedded to have forward compatible implementations.
type noopRefreshTokenUsecase struct{}

func (*noopRefreshTokenUsecase) GrantAuthenticatedUser(ctx context.Context, user *user.User) (*refreshtoken.RefreshToken, error) {
	panic("GrantAuthenticatedUser not implemented")
}

func (*noopRefreshTokenUsecase) GrantRefreshTokenSwap(ctx context.Context, user *user.User, token *refreshtoken.RefreshToken) (*refreshtoken.RefreshToken, error) {
	panic("GrantRefreshTokenSwap not implemented")
}

func (*noopRefreshTokenUsecase) FindRefreshTokenByID(ctx context.Context, token string) (*refreshtoken.RefreshToken, error) {
	panic("FindRefreshTokenByID not implemented")
}

func (*noopRefreshTokenUsecase) FindRefreshTokenByToken(ctx context.Context, token string) (*refreshtoken.RefreshToken, error) {
	panic("FindRefreshTokenByToken not implemented")
}

func (*noopRefreshTokenUsecase) Revoke(ctx context.Context, token *refreshtoken.RefreshToken) error {
	panic("Revoke not implemented")
}

func (*noopRefreshTokenUsecase) Logout(ctx context.Context, user *user.User) error {
	panic("Logout not implemented")
}
