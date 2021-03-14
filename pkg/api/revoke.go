package api

import (
	"errors"
	"net/http"

	"github.com/zbiljic/authzy/pkg/database"
	"github.com/zbiljic/authzy/pkg/logger"
)

// RevocationHandler is used to revoke OAuth2 Tokens.
func (s *server) RevocationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	token := r.FormValue("token")
	tokenType := r.FormValue("token_type_hint")

	ctx = s.log.NewContext(ctx, logger.Fields{"token_type_hint": tokenType})

	refreshToken, err := s.refreshTokenUsecase.FindRefreshTokenByToken(ctx, token)
	if err != nil {
		s.log.WithContext(ctx).Warnf("find refresh token: %v", err)

		if errors.Is(err, database.ErrNotFound) {
			w.WriteHeader(http.StatusOK)
			return
		}

		s.handleError(w, r, internalServerError(err.Error()))
		return
	}

	err = s.refreshTokenUsecase.Revoke(ctx, refreshToken)
	if err != nil {
		s.log.WithContext(ctx).Errorf("revoke: %v", err)

		s.handleError(w, r, internalServerError(err.Error()))
		return
	}

	s.log.WithContext(ctx).Info("token revoked")

	w.WriteHeader(http.StatusOK)
}
