package api

import (
	"net/http"

	"github.com/lestrrat-go/jwx/jwt"
	"github.com/mitchellh/mapstructure"

	"github.com/zbiljic/authzy/pkg/domain/refreshtoken"
	"github.com/zbiljic/authzy/pkg/logger"
)

// IntrospectHandler is used to introspect OAuth2 Tokens.
func (s *server) IntrospectHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tokenParam := r.FormValue("token")
	tokenTypeParam := r.FormValue("token_type_hint")

	ctx = s.log.NewContext(ctx, logger.Fields{"token_type_hint": tokenTypeParam})

	switch tokenTypeParam {
	case "access_token":
		accessToken, err := s.jwtService.Parse([]byte(tokenParam))
		if err != nil {
			s.log.WithContext(ctx).Warnf("parse access token: %v", err)

			handleIntrospectError(w)
			return
		}

		s.handleAccessTokenIntrospect(w, accessToken)
		return
	case "refresh_token":
		refreshToken, err := s.refreshTokenUsecase.FindRefreshTokenByToken(ctx, tokenParam)
		if err != nil {
			s.log.WithContext(ctx).Warnf("find refresh token: %v", err)

			handleIntrospectError(w)
			return
		}

		s.handleRefreshTokenIntrospect(w, r, refreshToken)
		return
	default:
	}

	// try as access token
	accessToken, err := s.jwtService.Parse([]byte(tokenParam))
	if err != nil {
		s.log.WithContext(ctx).Infof("parse as access token: %v", err)
		// ignore
	}

	if accessToken != nil {
		s.handleAccessTokenIntrospect(w, accessToken)
		return
	}

	// try as refresh token
	refreshToken, err := s.refreshTokenUsecase.FindRefreshTokenByToken(ctx, tokenParam)
	if err != nil {
		s.log.WithContext(ctx).Infof("find as refresh token: %v", err)
		// ignore
	}

	if refreshToken != nil {
		s.handleRefreshTokenIntrospect(w, r, refreshToken)
		return
	}

	handleIntrospectError(w)
}

func handleIntrospectError(w http.ResponseWriter) {
	resp := &Introspection{
		Active:   false,
		Audience: []string{},
	}
	mustSendJSON(w, http.StatusOK, resp)
}

func (s *server) handleAccessTokenIntrospect(w http.ResponseWriter, accessToken jwt.Token) {
	isActive := s.jwtService.Validate(accessToken) == nil

	resp := &Introspection{
		Active:    isActive,
		Subject:   accessToken.Subject(),
		ExpiresAt: accessToken.Expiration().Unix(),
		IssuedAt:  accessToken.IssuedAt().Unix(),
		NotBefore: accessToken.NotBefore().Unix(),
		Audience:  accessToken.Audience(),
		Issuer:    accessToken.Issuer(),
		TokenType: "bearer",
		TokenUse:  "access_token",
	}

	if resp.Audience == nil {
		resp.Audience = []string{}
	}

	if claim, ok := accessToken.Get(s.config.API.JWT.ClaimsNamespace); ok {
		var customClaims CustomClaims
		// ignoring error
		if err := mapstructure.Decode(claim, &customClaims); err == nil {
			resp.Username = customClaims.Username
			resp.Extra = make(map[string]interface{})
			resp.Extra["email"] = customClaims.Email
		}
	}

	mustSendJSON(w, http.StatusOK, resp)
}

func (s *server) handleRefreshTokenIntrospect(w http.ResponseWriter, r *http.Request, refreshToken *refreshtoken.RefreshToken) {
	ctx := r.Context()

	user, err := s.userUsecase.FindUserByID(ctx, refreshToken.UserID)
	if err != nil {
		handleIntrospectError(w)
		return
	}

	jwtToken := *getToken(ctx)

	sub := jwtToken.Subject()
	if sub == "" {
		handleIntrospectError(w)
		return
	}

	// user can only introspect its own tokens
	if user.ID != sub {
		handleIntrospectError(w)
		return
	}

	resp := &Introspection{
		Active:    !refreshToken.Revoked,
		Subject:   user.ID,
		IssuedAt:  refreshToken.CreatedAt.Unix(),
		NotBefore: refreshToken.CreatedAt.Unix(),
		Username:  user.Username,
		Audience:  []string{},
		TokenType: "bearer",
		TokenUse:  "refresh_token",
		Extra: map[string]interface{}{
			"email": user.Email,
		},
	}

	mustSendJSON(w, http.StatusOK, resp)
}
