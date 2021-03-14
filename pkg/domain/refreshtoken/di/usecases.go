package di

import (
	"go.uber.org/fx"

	"github.com/zbiljic/authzy/pkg/domain/refreshtoken"
	"github.com/zbiljic/authzy/pkg/domain/refreshtoken/usecases"
)

var usecasesfx = fx.Provide(
	NewRefreshTokenUsecase,
)

func NewRefreshTokenUsecase(
	repository refreshtoken.RefreshTokenRepository,
) refreshtoken.RefreshTokenUsecase {
	uc := usecases.NewRefreshTokenUsecase(
		repository,
	)
	return uc
}
