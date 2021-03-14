package usecases

import (
	"context"
	"errors"

	"github.com/zbiljic/authzy/pkg/database"
	"github.com/zbiljic/authzy/pkg/domain/refreshtoken"
	"github.com/zbiljic/authzy/pkg/domain/user"
	"github.com/zbiljic/authzy/pkg/ulid"
)

type refreshTokenUsecase struct {
	noopRefreshTokenUsecase

	repository refreshtoken.RefreshTokenRepository
}

func NewRefreshTokenUsecase(
	repository refreshtoken.RefreshTokenRepository,
) refreshtoken.RefreshTokenUsecase {
	uc := &refreshTokenUsecase{
		repository: repository,
	}
	return uc
}

func (uc *refreshTokenUsecase) GrantAuthenticatedUser(ctx context.Context, user *user.User) (*refreshtoken.RefreshToken, error) {
	token := &refreshtoken.RefreshToken{
		ID:     ulid.QuickULID().String(),
		UserID: user.ID,
		Token:  ulid.ULID().String(),
	}

	return uc.repository.Save(ctx, token)
}

func (uc *refreshTokenUsecase) GrantRefreshTokenSwap(ctx context.Context, user *user.User, token *refreshtoken.RefreshToken) (*refreshtoken.RefreshToken, error) {
	token.Revoked = true

	_, err := uc.repository.Save(ctx, token)
	if err != nil {
		return nil, err
	}

	return uc.GrantAuthenticatedUser(ctx, user)
}

func (uc *refreshTokenUsecase) FindRefreshTokenByID(ctx context.Context, id string) (*refreshtoken.RefreshToken, error) {
	return uc.repository.FindByID(ctx, id)
}

func (uc *refreshTokenUsecase) FindRefreshTokenByToken(ctx context.Context, token string) (*refreshtoken.RefreshToken, error) {
	return uc.repository.FindByToken(ctx, token)
}

func (uc *refreshTokenUsecase) Revoke(ctx context.Context, token *refreshtoken.RefreshToken) error {
	token.Revoked = true

	_, err := uc.repository.Save(ctx, token)
	if err != nil {
		// ignore missing tokens
		if errors.Is(err, database.ErrNotFound) {
			return nil
		}
	}

	return err
}

func (uc *refreshTokenUsecase) Logout(ctx context.Context, user *user.User) error {
	var (
		tokens     []*refreshtoken.RefreshToken
		nextCursor string
		err        error
	)

	for {
		tokens, nextCursor, err = uc.repository.FindAllForUser(ctx, user.ID, nextCursor, 0)
		if err != nil {
			return err
		}

		for _, refreshToken := range tokens {
			err = uc.repository.Delete(ctx, refreshToken)
			if err != nil {
				return err
			}
		}

		if len(tokens) == 0 {
			break
		}
	}

	return nil
}
