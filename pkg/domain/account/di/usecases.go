package di

import (
	"go.uber.org/fx"

	"github.com/zbiljic/authzy/pkg/domain/account"
	"github.com/zbiljic/authzy/pkg/domain/account/usecases"
)

var usecasesfx = fx.Provide(
	NewAccountUsecase,
)

func NewAccountUsecase(
	repository account.AccountRepository,
) account.AccountUsecase {
	uc := usecases.NewAccountUsecase(
		repository,
	)
	return uc
}
