package api

import (
	"context"
	"net/http"
	"regexp"

	"github.com/lestrrat-go/jwx/jwt"

	xhttp "github.com/zbiljic/authzy/pkg/http"
)

var bearerRegexp = regexp.MustCompile(`^(?:B|b)earer (\S+$)`)

func (s *server) AuthHandler(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tokenString, err := s.extractBearerToken(r)
		if err != nil {
			s.log.WithContext(ctx).Errorf("extract bearer token: %v", err)

			s.clearCookieToken(ctx, w)
			s.handleError(w, r, err)
			return
		}

		jwtToken, err := s.parseJWT(ctx, tokenString)
		if err != nil {
			s.log.WithContext(ctx).Errorf("parse JWT token: %v", err)

			s.clearCookieToken(ctx, w)
			s.handleError(w, r, err)
			return
		}

		ctx = withToken(ctx, &jwtToken)

		next(w, r.WithContext(ctx))
	})
}

func (s *server) extractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get(xhttp.Authorization)
	if authHeader == "" {
		return "", unauthorizedError("This endpoint requires a Bearer token")
	}

	matches := bearerRegexp.FindStringSubmatch(authHeader)
	if len(matches) != 2 {
		return "", unauthorizedError("This endpoint requires a Bearer token")
	}

	return matches[1], nil
}

func (s *server) parseJWT(ctx context.Context, bearer string) (jwt.Token, error) {
	jwtToken, err := s.jwtService.Parse([]byte(bearer))
	if err != nil {
		return nil, unauthorizedError("Invalid token: %v", err)
	}

	err = s.jwtService.Validate(jwtToken)
	if err != nil {
		return nil, unauthorizedError("Invalid token: %v", err)
	}

	return jwtToken, nil
}
