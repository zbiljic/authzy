package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/zbiljic/authzy/pkg/database"
	"github.com/zbiljic/authzy/pkg/domain/user"
	xhttp "github.com/zbiljic/authzy/pkg/http"
	"github.com/zbiljic/authzy/pkg/logger"
)

var (
	// used below to specify need to return answer to user via specific redirect
	errRedirectWithQuery = errors.New("need return answer with query params")
)

const (
	signupVerification   = "signup"
	recoveryVerification = "recovery"
)

// VerifyHandler exchanges a confirmation or recovery token to a refresh token.
func (s *server) VerifyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := &VerifyRequest{}
	cookie := r.Header.Get(xhttp.XUseCookie)

	switch r.Method {
	// GET only supports signup type
	case http.MethodGet:
		params.Type = r.FormValue("type")
		params.Token = r.FormValue("token")
		params.Password = ""
		params.RedirectTo = s.validateRedirectURL(r, r.FormValue("redirect_to"))
	case http.MethodPost:
		jsonDecoder := json.NewDecoder(r.Body)
		if err := jsonDecoder.Decode(params); err != nil {
			s.handleError(w, r, badRequestError("Could not read verification params: %v", err))
			return
		}
	default:
		s.handleError(w, r, unprocessableEntityError("Only GET and POST methods are supported"))
		return
	}

	if params.Token == "" {
		s.handleError(w, r, unprocessableEntityError("Verify requires a token"))
		return
	}

	ctx = s.log.NewContext(ctx, logger.Fields{"verify_type": params.Type})

	var (
		user  *user.User
		err   error
		token *AccessTokenResponse
	)

	switch params.Type {
	case signupVerification:
		user, err = s.signupVerify(ctx, params)
	case recoveryVerification:
		user, err = s.recoverVerify(ctx, params)
	default:
		s.handleError(w, r, unprocessableEntityError("Verify requires a verification type"))
		return
	}

	if err != nil {
		var e *HTTPError
		if errors.As(err, &e) {
			if errors.Is(e.InternalError, errRedirectWithQuery) {
				rURL := s.prepErrorRedirectURL(e, r)

				s.log.WithContext(ctx).Errorf("error redirect: %v", err)

				http.Redirect(w, r, rURL, http.StatusFound)
				return
			}
		}

		s.handleError(w, r, err)
		return
	}

	ctx = s.log.NewContext(ctx, logger.Fields{"user_id": user.ID})

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

	s.log.WithContext(ctx).Info("user verification completed")

	if cookie != "" && s.config.API.Cookie.DurationSeconds > 0 {
		err := s.setCookieToken(ctx, w, token.Token, cookie == useSessionCookie)
		if err != nil {
			s.log.WithContext(ctx).Errorf("set cookie: %v", err)

			s.handleError(w, r, internalServerError("Failed to set JWT cookie. %s", err))
			return
		}
	}

	switch r.Method {
	// GET requests should return to the app site after confirmation
	case http.MethodGet:
		rURL := params.RedirectTo
		if rURL == "" {
			rURL = s.config.SiteURL
		}
		if token != nil {
			q := url.Values{}
			q.Set("access_token", token.Token)
			q.Set("token_type", token.TokenType)
			q.Set("expires_in", strconv.Itoa(token.ExpiresIn))
			q.Set("refresh_token", token.RefreshToken)
			q.Set("type", params.Type)
			rURL += "#" + q.Encode()
		}
		http.Redirect(w, r, rURL, http.StatusSeeOther)
	case http.MethodPost:
		mustSendJSON(w, http.StatusOK, token)
		return
	}
}

func (s *server) signupVerify(ctx context.Context, params *VerifyRequest) (*user.User, error) {
	user, err := s.userUsecase.FindUserByConfirmationToken(ctx, params.Token)
	if err != nil {
		s.log.WithContext(ctx).Warnf("find user: %v", err)

		if errors.Is(err, database.ErrNotFound) {
			return nil, notFoundError(err.Error())
		}

		return nil, internalServerError("Database error finding user").WithInternalError(err)
	}

	nextDay := user.ConfirmationSentAt.Add(24 * time.Hour)
	if user.ConfirmationSentAt != nil && time.Now().After(nextDay) {
		return nil, goneError("Confirmation token expired")
	}

	user, err = s.userUsecase.ConfirmUser(ctx, user.ID)
	if err != nil {
		return nil, internalServerError("Error confirming user").WithInternalError(err)
	}

	return user, nil
}

func (s *server) recoverVerify(ctx context.Context, params *VerifyRequest) (*user.User, error) {
	user, err := s.userUsecase.FindUserByRecoveryToken(ctx, params.Token)
	if err != nil {
		s.log.WithContext(ctx).Warnf("find user: %v", err)

		if errors.Is(err, database.ErrNotFound) {
			return nil, notFoundError(err.Error())
		}

		return nil, internalServerError("Database error finding user").WithInternalError(err)
	}

	nextDay := user.RecoverySentAt.Add(24 * time.Hour)
	if user.RecoverySentAt != nil && time.Now().After(nextDay) {
		return nil, goneError("Recovery token expired")
	}

	user, err = s.userUsecase.ConfirmRecovery(ctx, user)
	if err != nil {
		return nil, internalServerError("Error confirming recovery").WithInternalError(err)
	}

	return user, nil
}

func (s *server) prepErrorRedirectURL(err *HTTPError, r *http.Request) string {
	rURL := s.config.SiteURL

	q := url.Values{}
	if str, ok := oauthErrorMap[err.Code]; ok {
		q.Set("error", str)
	}
	q.Set("error_code", strconv.Itoa(err.Code))
	q.Set("error_description", err.Message)

	return rURL + "#" + q.Encode()
}
