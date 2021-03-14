package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hako/durafmt"
	"github.com/zbiljic/authzy/pkg/database"
	"github.com/zbiljic/authzy/pkg/domain/account"
	"github.com/zbiljic/authzy/pkg/domain/user"
	"github.com/zbiljic/authzy/pkg/logger"
)

// SignupHandler is the endpoint for registering a new user.
func (s *server) SignupHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.ContentLength == 0 {
		s.handleError(w, r, badRequestError("Empty request body"))
		return
	}

	params := &SignupRequest{}

	jsonDecoder := json.NewDecoder(r.Body)
	err := jsonDecoder.Decode(params)
	if err != nil {
		s.handleError(w, r, badRequestError("Could not read input params: %v", err))
		return
	}

	if params.Email == "" {
		s.handleError(w, r, unprocessableEntityError("Account creation requires a valid email"))
		return
	}
	if params.Password == "" {
		s.handleError(w, r, unprocessableEntityError("Account creation requires a valid password"))
		return
	}
	if err := s.validateEmail(ctx, params.Email); err != nil {
		s.handleError(w, r, err)
		return
	}

	createdUser, err := s.userUsecase.FindUserByEmail(ctx, params.Email)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			s.handleError(w, r, internalServerError("Database error finding user").WithInternalError(err))
			return
		}
	}

	if createdUser != nil {
		ctx = s.log.NewContext(ctx, logger.Fields{"user_id": createdUser.ID, "email": createdUser.Email, "username": createdUser.Username})

		if createdUser.IsConfirmed() {
			s.log.WithContext(ctx).Warn("already registered")

			s.handleError(w, r, badRequestError("A user with this email address has already been registered"))
			return
		}

		s.log.WithContext(ctx).Warn("user not confirmed")

		if _, err := s.userUsecase.UpdateUserMetaData(ctx, createdUser, params.UserMetaData); err != nil {
			s.handleError(w, r, internalServerError("Database error updating user").WithInternalError(err))
			return
		}

		s.log.WithContext(ctx).Info("user updated")
	} else {
		s.log.WithContext(ctx).
			WithFields(logger.Fields{"email": params.Email, "username": params.Username}).
			Info("creating user")

		createdUser, err = s.signupNewUser(ctx, params)
		if err != nil {
			s.log.WithContext(ctx).Errorf("could not create user: %v", err)

			s.handleError(w, r, internalServerError("Could not create user").WithInternalError(err))
			return
		}

		s.log.WithContext(ctx).
			WithFields(logger.Fields{"user_id": createdUser.ID, "email": createdUser.Email, "username": createdUser.Username}).
			Info("new user created")
	}

	if s.config.API.Mailer.Autoconfirm {
		_, err = s.userUsecase.ConfirmUser(ctx, createdUser.ID)
		if err != nil {
			s.log.WithContext(ctx).Errorf("could not confirm user: %v", err)

			s.handleError(w, r, internalServerError("Could not update user").WithInternalError(err))
			return
		}
	} else {
		mailer := s.Mailer(ctx)
		referrer := s.getReferrer(r)
		if err = s.sendConfirmation(ctx, createdUser, mailer, s.config.SMTP.MaxFrequency, referrer); err != nil {
			if errors.Is(err, ErrMaxFrequencyLimit) {
				maxFrequencyHumanString := durafmt.Parse(s.config.SMTP.MaxFrequency).String()
				s.handleError(w, r, tooManyRequestsError("For security purposes, you can only request this once every %s", maxFrequencyHumanString))
				return
			}

			s.handleError(w, r, internalServerError("Error sending confirmation mail").WithInternalError(err))
			return
		}
	}

	out := &SignupResponse{
		ID:            createdUser.ID,
		Email:         createdUser.Email,
		EmailVerified: createdUser.EmailVerified,
		Username:      createdUser.Username,
		GivenName:     createdUser.GivenName,
		FamilyName:    createdUser.FamilyName,
		Name:          createdUser.Name,
		Nickname:      createdUser.Nickname,
		Picture:       createdUser.Picture,
	}

	mustSendJSON(w, http.StatusCreated, out)
}

func (s *server) signupNewUser(ctx context.Context, in *SignupRequest) (*user.User, error) {
	createUserRequest := user.User{
		Email:      in.Email,
		Username:   in.Username,
		GivenName:  in.GivenName,
		FamilyName: in.FamilyName,
		Name:       in.Name,
		Nickname:   in.Nickname,
		Picture:    in.Picture,
	}

	createdUser, err := s.userUsecase.CreateUser(ctx, &createUserRequest)
	if err != nil {
		return nil, err
	}

	passwordAccount := &account.Account{
		UserID:      createdUser.ID,
		Provider:    account.ProviderTypePassword,
		FederatedID: createUserRequest.Email,
	}

	_, err = s.accountUsecase.CreateAccount(ctx, passwordAccount)
	if err != nil {
		return nil, err
	}

	createdUser, err = s.userUsecase.UpdatePassword(ctx, createdUser.ID, []byte(in.Password))
	if err != nil {
		return nil, err
	}

	return createdUser, nil
}
