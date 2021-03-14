package api

import (
	"context"
	"net/http"
	"time"
)

const useSessionCookie = "session"

func (s *server) setCookieToken(ctx context.Context, w http.ResponseWriter, tokenString string, session bool) error {
	config := getConfig(ctx)

	cookie := &http.Cookie{
		Name:     config.API.Cookie.Key,
		Value:    tokenString,
		Secure:   config.API.Secure,
		HttpOnly: true,
		Path:     "/",
	}

	if !session {
		exp := time.Second * time.Duration(config.API.Cookie.DurationSeconds)

		cookie.Expires = time.Now().Add(exp)
		cookie.MaxAge = config.API.Cookie.DurationSeconds
	}

	http.SetCookie(w, cookie)
	return nil
}

func (s *server) clearCookieToken(ctx context.Context, w http.ResponseWriter) {
	config := getConfig(ctx)

	http.SetCookie(w, &http.Cookie{
		Name:     config.API.Cookie.Key,
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour * 10),
		MaxAge:   -1,
		Secure:   config.API.Secure,
		HttpOnly: true,
		Path:     "/",
	})
}
