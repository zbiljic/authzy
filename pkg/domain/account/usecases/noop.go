package usecases

import (
	"context"

	"github.com/zbiljic/authzy/pkg/domain/account"
)

// Compile-time proof of interface implementation.
var _ account.AccountUsecase = (*noopAccountUsecase)(nil)

// noopAccountUsecase can be embedded to have forward compatible implementations.
type noopAccountUsecase struct{}

func (*noopAccountUsecase) CreateAccount(ctx context.Context, account *account.Account) (*account.Account, error) {
	panic("CreateAccount not implemented")
}

func (*noopAccountUsecase) UpdateAccount(ctx context.Context, account *account.Account) (*account.Account, error) {
	panic("UpdateAccount not implemented")
}

func (*noopAccountUsecase) FindAllForUser(ctx context.Context, userID string) ([]*account.Account, error) {
	panic("FindAllForUser not implemented")
}
