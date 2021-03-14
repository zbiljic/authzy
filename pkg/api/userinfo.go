package api

import (
	"errors"
	"net/http"

	"github.com/zbiljic/authzy/pkg/database"
	"github.com/zbiljic/authzy/pkg/logger"
)

// UserinfoHandler returns the payload of the ID Token of the provided
// OAuth 2.0 Access Token.
func (s *server) UserinfoHandler(w http.ResponseWriter, r *http.Request) {
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

	resp := &UserinfoResponse{
		Sub:               sub,
		Name:              user.Name,
		GivenName:         user.GivenName,
		FamilyName:        user.FamilyName,
		Nickname:          user.Nickname,
		PreferredUsername: user.Username,
		Picture:           user.Picture,
		Email:             user.Email,
		EmailVerified:     user.EmailVerified,
		UpdatedAt:         user.UpdatedAt.Unix(),
	}

	mustSendJSON(w, http.StatusOK, resp)
}
