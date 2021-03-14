package usecases

import (
	"context"

	"github.com/zbiljic/authzy/pkg/domain/account"
)

type accountUsecase struct {
	noopAccountUsecase

	repository account.AccountRepository
}

func NewAccountUsecase(
	repository account.AccountRepository,
) account.AccountUsecase {
	uc := &accountUsecase{
		repository: repository,
	}
	return uc
}

func (uc *accountUsecase) CreateAccount(ctx context.Context, account *account.Account) (*account.Account, error) {
	if err := account.Provider.IsValid(); err != nil {
		return nil, err
	}
	return uc.repository.Save(ctx, account)
}

func (uc *accountUsecase) UpdateAccount(ctx context.Context, account *account.Account) (*account.Account, error) {
	if err := account.Provider.IsValid(); err != nil {
		return nil, err
	}
	return uc.repository.Save(ctx, account)
}

func (uc *accountUsecase) FindAllForUser(ctx context.Context, userID string) ([]*account.Account, error) {
	return uc.repository.FindAllForUser(ctx, userID)
}
