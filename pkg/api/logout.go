package api

import (
	"net/http"
	"strings"
)

// LogoutHandler is the endpoint for logging out a user and thereby revoking
// any refresh tokens.
func (s *server) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	redirectTo := r.FormValue("redirect_to")

	s.clearCookieToken(ctx, w)

	user, err := s.getUserFromToken(ctx)
	if err != nil {
		s.handleError(w, r, unauthorizedError("Invalid user").WithInternalError(err))
		return
	}

	err = s.refreshTokenUsecase.Logout(ctx, user)
	if err != nil {
		s.handleError(w, r, internalServerError("Error logging out user").WithInternalError(err))
		return
	}

	if redirectTo != "" {
		for _, allowedLogoutURL := range s.config.API.AllowedLogoutURLs {
			if strings.EqualFold(allowedLogoutURL, redirectTo) {
				http.Redirect(w, r, redirectTo, http.StatusSeeOther)
				return
			}
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
