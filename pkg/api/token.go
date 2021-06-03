package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/zbiljic/authzy/pkg/database"
	"github.com/zbiljic/authzy/pkg/domain/user"
	xhttp "github.com/zbiljic/authzy/pkg/http"
	"github.com/zbiljic/authzy/pkg/logger"
)

// TokenHandler is the endpoint for OAuth access token requests.
func (s *server) TokenHandler(w http.ResponseWriter, r *http.Request) {
	grantType := r.FormValue("grant_type")

	switch grantType {
	case "password":
		s.ResourceOwnerPasswordGrant(w, r)
	case "refresh_token":
		s.RefreshTokenGrant(w, r)
	default:
		s.handleError(w, r, oauthError("unsupported_grant_type", ""))
	}
}

// ResourceOwnerPasswordGrant implements the password grant type flow.
func (s *server) ResourceOwnerPasswordGrant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	username := r.FormValue("username")
	password := r.FormValue("password")
	cookie := r.Header.Get(xhttp.XUseCookie)

	if username == "" {
		s.log.WithContext(ctx).Warn("username required")

		s.handleError(w, r, oauthError("invalid_grant", "Field username required."))
		return
	}
	if password == "" {
		s.log.WithContext(ctx).Warn("password required")

		s.handleError(w, r, oauthError("invalid_grant", "Field password required."))
		return
	}

	user, err := s.userUsecase.Authenticate(ctx, username, []byte(password))
	if err != nil {
		s.log.WithContext(ctx).
			WithFields(logger.Fields{"identifier": username}).
			Warnf("authentication failed: %v", err)

		s.handleError(w, r, oauthError("invalid_grant", "No user found with that identifier, or password invalid."))
		return
	}

	ctx = s.log.NewContext(ctx, logger.Fields{"user_id": user.ID})

	if !user.IsConfirmed() {
		s.log.WithContext(ctx).Warn("email not confirmed")

		s.handleError(w, r, oauthError("invalid_grant", "Email not confirmed"))
		return
	}

	var token *AccessTokenResponse

	token, err = s.issueRefreshToken(ctx, user)
	if err != nil {
		if e, ok := err.(ErrorCause); ok {
			s.log.WithContext(ctx).Errorf("issue refresh token: %v", e.Cause())
		} else {
			s.log.WithContext(ctx).Errorf("issue refresh token: %v", err)
		}

		s.handleError(w, r, internalServerError("Failed to issue refresh token. %s", err))
		return
	}

	s.log.WithContext(ctx).Info("issued refresh token")

	if cookie != "" && s.config.API.Cookie.DurationSeconds > 0 {
		err := s.setCookieToken(ctx, w, token.Token, cookie == useSessionCookie)
		if err != nil {
			s.log.WithContext(ctx).Errorf("set cookie: %v", err)

			s.handleError(w, r, internalServerError("Failed to set JWT cookie. %s", err))
			return
		}
	}

	mustSendJSON(w, http.StatusOK, token)
}

// RefreshTokenGrant implements the refresh_token grant type flow.
func (s *server) RefreshTokenGrant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	refreshTokenParam := r.FormValue("refresh_token")
	cookie := r.Header.Get(xhttp.XUseCookie)

	if refreshTokenParam == "" {
		s.log.WithContext(ctx).Warn("refresh_token required")

		s.handleError(w, r, oauthError("invalid_request", "refresh_token required"))
		return
	}

	token, err := s.refreshTokenUsecase.FindRefreshTokenByToken(ctx, refreshTokenParam)
	if err != nil {
		s.log.WithContext(ctx).
			WithFields(logger.Fields{"refresh_token": refreshTokenParam}).
			Warnf("find refresh token: %v", err)

		if errors.Is(err, database.ErrNotFound) {
			s.handleError(w, r, oauthError("invalid_request", "Invalid Refresh Token"))
			return
		}

		s.handleError(w, r, internalServerError(err.Error()))
		return
	}

	user, err := s.userUsecase.FindUserByID(ctx, token.UserID)
	if err != nil {
		s.log.WithContext(ctx).
			WithFields(logger.Fields{"user_id": token.UserID}).
			Warnf("find user: %v", err)

		if errors.Is(err, database.ErrNotFound) {
			s.handleError(w, r, oauthError("invalid_request", "Invalid Refresh Token"))
			return
		}

		s.handleError(w, r, internalServerError(err.Error()))
		return
	}

	ctx = s.log.NewContext(ctx, logger.Fields{"user_id": user.ID})

	if token.Revoked {
		s.log.WithContext(ctx).
			WithFields(logger.Fields{"refresh_token": refreshTokenParam}).
			Error("refresh token revoked")

		s.clearCookieToken(ctx, w)

		err := oauthError("invalid_grant", "Invalid Refresh Token").
			WithInternalMessage("Possible abuse attempt: %v", r)
		s.handleError(w, r, err)
		return
	}

	newToken, err := s.refreshTokenUsecase.GrantRefreshTokenSwap(ctx, user, token)
	if err != nil {
		s.log.WithContext(ctx).Errorf("swap refresh token: %v", err)

		s.handleError(w, r, internalServerError("Failed to swap refresh token. %s", err))
		return
	}

	tokenString, err := s.generateAccessToken(ctx, user, s.config.API.JWT.ClaimsNamespace)
	if err != nil {
		s.log.WithContext(ctx).Errorf("generate access token: %v", err)

		s.handleError(w, r, internalServerError("error generating jwt token").WithInternalError(err))
		return
	}

	s.log.WithContext(ctx).Info("refreshed token")

	if cookie != "" && s.config.API.Cookie.DurationSeconds > 0 {
		err := s.setCookieToken(ctx, w, tokenString, cookie == useSessionCookie)
		if err != nil {
			s.log.WithContext(ctx).Errorf("set cookie: %v", err)

			s.handleError(w, r, internalServerError("Failed to set JWT cookie. %s", err))
			return
		}
	}

	resp := &AccessTokenResponse{
		Token:        tokenString,
		TokenType:    "bearer",
		ExpiresIn:    s.config.API.JWT.Exp,
		RefreshToken: newToken.Token,
	}

	mustSendJSON(w, http.StatusOK, resp)
}

func (s *server) issueRefreshToken(ctx context.Context, user *user.User) (*AccessTokenResponse, error) {
	refreshToken, err := s.refreshTokenUsecase.GrantAuthenticatedUser(ctx, user)
	if err != nil {
		return nil, internalServerError("error granting user").WithInternalError(err)
	}

	tokenString, err := s.generateAccessToken(ctx, user, s.config.API.JWT.ClaimsNamespace)
	if err != nil {
		return nil, internalServerError("error generating jwt token").WithInternalError(err)
	}

	userIP := getUserIP(ctx)

	_, err = s.userUsecase.UserSignedIn(ctx, user, userIP)
	if err != nil {
		return nil, internalServerError("error updating user").WithInternalError(err)
	}

	return &AccessTokenResponse{
		Token:        tokenString,
		TokenType:    "bearer",
		ExpiresIn:    s.config.API.JWT.Exp,
		RefreshToken: refreshToken.Token,
	}, nil
}

func (s *server) generateAccessToken(ctx context.Context, user *user.User, claimsNamespace string) (string, error) {
	token, err := s.jwtService.Generate(user.ID)
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	// add custom claims
	customClaims := &CustomClaims{
		Username: user.Username,
		Email:    user.Email,
	}

	err = token.Set(claimsNamespace, customClaims)
	if err != nil {
		return "", fmt.Errorf("set custom claims: %w", err)
	}

	signed, err := s.jwtService.Sign(token)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}
