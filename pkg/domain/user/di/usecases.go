package di

import (
	"go.uber.org/fx"

	"github.com/zbiljic/authzy/pkg/domain/user"
	"github.com/zbiljic/authzy/pkg/domain/user/usecases"
	"github.com/zbiljic/authzy/pkg/hash"
)

var usecasesfx = fx.Provide(
	NewUserUsecase,
)

func NewUserUsecase(
	hasher hash.Hasher,
	repository user.UserRepository,
) user.UserUsecase {
	uc := usecases.NewUserUsecase(
		hasher,
		repository,
	)
	return uc
}
