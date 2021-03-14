package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hako/durafmt"
	"github.com/zbiljic/authzy/pkg/database"
)

// RecoverHandler sends a recovery email.
func (s *server) RecoverHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.ContentLength == 0 {
		s.handleError(w, r, badRequestError("Empty request body"))
		return
	}

	params := &RecoverRequest{}

	jsonDecoder := json.NewDecoder(r.Body)
	err := jsonDecoder.Decode(params)
	if err != nil {
		s.handleError(w, r, badRequestError("Could not read input params: %v", err))
		return
	}

	if params.Email == "" {
		s.handleError(w, r, unprocessableEntityError("Password recovery requires an email"))
		return
	}
	if err := s.validateEmail(ctx, params.Email); err != nil {
		s.handleError(w, r, err)
		return
	}

	user, err := s.userUsecase.FindUserByEmail(ctx, params.Email)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			s.handleError(w, r, notFoundError(err.Error()))
			return
		}

		s.handleError(w, r, internalServerError("Database error finding user").WithInternalError(err))
		return
	}

	mailer := s.Mailer(ctx)
	referrer := s.getReferrer(r)
	if err = s.sendPasswordRecovery(ctx, user, mailer, s.config.SMTP.MaxFrequency, referrer); err != nil {
		if errors.Is(err, ErrMaxFrequencyLimit) {
			maxFrequencyHumanString := durafmt.Parse(s.config.SMTP.MaxFrequency).String()
			s.handleError(w, r, tooManyRequestsError("For security purposes, you can only request this once every %s", maxFrequencyHumanString))
			return
		}

		s.handleError(w, r, internalServerError("Error recovering user").WithInternalError(err))
		return
	}

	mustSendJSON(w, http.StatusOK, &map[string]string{})
}
