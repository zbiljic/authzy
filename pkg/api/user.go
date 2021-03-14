package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/zbiljic/authzy/pkg/database"
	"github.com/zbiljic/authzy/pkg/logger"
)

func (s *server) UserGetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	jwtToken := *getToken(ctx)

	sub := jwtToken.Subject()
	if sub == "" {
		s.handleError(w, r, badRequestError("Could not read 'sub' claim"))
		return
	}

	ctx = s.log.NewContext(ctx, logger.Fields{"user_id": sub})

	user, err := s.userUsecase.FindUserByID(ctx, sub)
	if err != nil {
		s.log.WithContext(ctx).Warnf("find user: %v", err)

		if errors.Is(err, database.ErrNotFound) {
			s.handleError(w, r, notFoundError(err.Error()))
			return
		}

		s.handleError(w, r, internalServerError(err.Error()))
		return
	}

	accounts, err := s.accountUsecase.FindAllForUser(ctx, sub)
	if err != nil {
		s.log.WithContext(ctx).Warnf("find accounts: %v", err)

		s.handleError(w, r, internalServerError(err.Error()))
		return
	}

	resp := &UserResponse{
		UserID:        user.ID,
		Email:         user.Email,
		EmailVerified: user.IsConfirmed(),
		Username:      user.Username,
		GivenName:     user.GivenName,
		FamilyName:    user.FamilyName,
		Name:          user.Name,
		Nickname:      user.Nickname,
		Picture:       user.Picture,
		AppMetaData:   user.AppMetaData,
		UserMetaData:  user.UserMetaData,
		Providers:     make([]Provider, 0),
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}

	for _, acc := range accounts {
		provider := Provider{
			Provider:    acc.Provider.String(),
			FederatedID: acc.FederatedID,
		}
		resp.Providers = append(resp.Providers, provider)
	}

	mustSendJSON(w, http.StatusOK, resp)
}

func (s *server) UserUpdateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := &UserUpdateRequest{}

	jsonDecoder := json.NewDecoder(r.Body)
	err := jsonDecoder.Decode(params)
	if err != nil {
		s.handleError(w, r, badRequestError("Could not read input params: %v", err))
		return
	}

	jwtToken := *getToken(ctx)

	sub := jwtToken.Subject()
	if sub == "" {
		s.handleError(w, r, badRequestError("Could not read 'sub' claim"))
		return
	}

	ctx = s.log.NewContext(ctx, logger.Fields{"user_id": sub})

	user, err := s.userUsecase.FindUserByID(ctx, sub)
	if err != nil {
		s.log.WithContext(ctx).Warnf("find user: %v", err)

		if errors.Is(err, database.ErrNotFound) {
			s.handleError(w, r, notFoundError(err.Error()))
			return
		}

		s.handleError(w, r, internalServerError(err.Error()))
		return
	}

	if params.Password != "" {
		user, err = s.userUsecase.UpdatePassword(ctx, user.ID, []byte(params.Password))
		if err != nil {
			s.handleError(w, r, internalServerError("Error during password storage").WithInternalError(err))
			return
		}
	}

	if params.UserMetaData != nil {
		user, err = s.userUsecase.UpdateUserMetaData(ctx, user, params.UserMetaData)
		if err != nil {
			s.handleError(w, r, internalServerError("Error updating user").WithInternalError(err))
			return
		}
	}

	if params.EmailChangeToken != "" {
		s.log.WithContext(ctx).Debugf("email change token %v", params.EmailChangeToken)

		if params.EmailChangeToken != user.EmailChangeToken {
			s.handleError(w, r, unauthorizedError("email change token invalid"))
			return
		}

		user, err = s.userUsecase.ConfirmEmailChange(ctx, user)
		if err != nil {
			s.handleError(w, r, internalServerError("Error updating user").WithInternalError(err))
			return
		}
	} else if params.Email != "" && params.Email != user.Email {
		if err := s.validateEmail(ctx, params.Email); err != nil {
			s.handleError(w, r, err)
			return
		}

		var exists bool
		emailUser, err := s.userUsecase.FindUserByEmail(ctx, params.Email)
		if err != nil {
			if !errors.Is(err, database.ErrNotFound) {
				s.handleError(w, r, internalServerError("Database error finding user").WithInternalError(err))
				return
			}
		}

		if emailUser != nil {
			exists = true
		}

		if exists {
			s.handleError(w, r, unprocessableEntityError("Email address already registered by another user"))
			return
		}

		mailer := s.Mailer(ctx)
		referrer := s.getReferrer(r)
		if err = s.sendEmailChange(ctx, user, mailer, params.Email, referrer); err != nil {
			s.handleError(w, r, internalServerError("Error sending change email").WithInternalError(err))
			return
		}
	}

	s.log.WithContext(ctx).Info("user updated")

	accounts, err := s.accountUsecase.FindAllForUser(ctx, sub)
	if err != nil {
		s.log.WithContext(ctx).Warnf("find accounts: %v", err)

		s.handleError(w, r, internalServerError(err.Error()))
		return
	}

	resp := &UserResponse{
		UserID:        user.ID,
		Email:         user.Email,
		EmailVerified: user.IsConfirmed(),
		Username:      user.Username,
		GivenName:     user.GivenName,
		FamilyName:    user.FamilyName,
		Name:          user.Name,
		Nickname:      user.Nickname,
		Picture:       user.Picture,
		AppMetaData:   user.AppMetaData,
		UserMetaData:  user.UserMetaData,
		Providers:     make([]Provider, 0),
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}

	for _, acc := range accounts {
		provider := Provider{
			Provider:    acc.Provider.String(),
			FederatedID: acc.FederatedID,
		}
		resp.Providers = append(resp.Providers, provider)
	}

	mustSendJSON(w, http.StatusOK, resp)
}
