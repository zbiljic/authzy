package di

import (
	"context"
	"net/http"

	"go.uber.org/fx"

	"github.com/zbiljic/authzy/pkg/api"
	"github.com/zbiljic/authzy/pkg/config"
	"github.com/zbiljic/authzy/pkg/domain/account"
	"github.com/zbiljic/authzy/pkg/domain/refreshtoken"
	"github.com/zbiljic/authzy/pkg/domain/user"
	"github.com/zbiljic/authzy/pkg/jwt"
	"github.com/zbiljic/authzy/pkg/logger"
)

var apifx = fx.Provide(APIHandlerProvider)

type APIHandlerParams struct {
	fx.In

	Lifecycle fx.Lifecycle

	Log    logger.Logger
	Config *config.Config

	JWTService          jwt.Service
	AccountUsecase      account.AccountUsecase
	RefreshTokenUsecase refreshtoken.RefreshTokenUsecase
	UserUsecase         user.UserUsecase
}

type APIHandlerResult struct {
	fx.Out

	Router http.Handler `name:"api"`
}

func APIHandlerProvider(p APIHandlerParams) (APIHandlerResult, error) {
	apiService := api.New(
		p.Log,
		p.Config,
		p.JWTService,
		p.AccountUsecase,
		p.RefreshTokenUsecase,
		p.UserUsecase,
	)

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			p.Log.Debug("starting API service")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			p.Log.Debug("stopping API service")
			return apiService.Close()
		},
	})

	return APIHandlerResult{Router: apiService}, nil
}
