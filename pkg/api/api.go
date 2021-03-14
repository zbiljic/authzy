package api

import (
	"context"
	"io"
	"net/http"

	"github.com/zbiljic/authzy/pkg/config"
	"github.com/zbiljic/authzy/pkg/domain/account"
	"github.com/zbiljic/authzy/pkg/domain/refreshtoken"
	"github.com/zbiljic/authzy/pkg/domain/user"
	"github.com/zbiljic/authzy/pkg/jwt"
	"github.com/zbiljic/authzy/pkg/logger"
	"github.com/zbiljic/authzy/pkg/mailer"
)

// SlashSeparator - slash separator.
const SlashSeparator = "/"

// Service is the API service interface.
type Service interface {
	http.Handler
	io.Closer
}

type server struct {
	http.Handler

	log    logger.Logger
	config *config.Config

	jwtService          jwt.Service
	accountUsecase      account.AccountUsecase
	refreshTokenUsecase refreshtoken.RefreshTokenUsecase
	userUsecase         user.UserUsecase
}

// New will create a and initialize a new API service.
func New(
	log logger.Logger,
	config *config.Config,
	jwtService jwt.Service,
	accountUsecase account.AccountUsecase,
	refreshTokenUsecase refreshtoken.RefreshTokenUsecase,
	userUsecase user.UserUsecase,
) Service {
	s := &server{
		log:                 log,
		config:              config,
		jwtService:          jwtService,
		accountUsecase:      accountUsecase,
		refreshTokenUsecase: refreshTokenUsecase,
		userUsecase:         userUsecase,
	}

	s.setupRouting()

	return s
}

func (s *server) Close() error {
	s.log.Info("api shutting down")
	return nil
}

func (s *server) handleError(w http.ResponseWriter, r *http.Request, err error) {
	handleError(w, r, s.log, err)
}

func (s *server) Mailer(ctx context.Context) mailer.Mailer {
	return mailer.NewMailer(s.log.WithContext(ctx), s.config)
}
