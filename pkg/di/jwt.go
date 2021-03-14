package di

import (
	"go.uber.org/fx"

	"github.com/zbiljic/authzy/pkg/config"
	"github.com/zbiljic/authzy/pkg/jwt"
)

var jwtfx = fx.Provide(
	ProvideJWTService,
)

func ProvideJWTService(config *config.JWTConfig) (jwt.Service, error) {
	return jwt.NewService(config)
}
