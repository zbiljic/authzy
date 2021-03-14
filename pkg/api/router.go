package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/zbiljic/authzy"
	"github.com/zbiljic/authzy/pkg/config"
	xhttp "github.com/zbiljic/authzy/pkg/http"
	"github.com/zbiljic/authzy/pkg/logger"
)

const (
	CSRFPath   = "/csrf"
	SignupPath = "/signup"

	TokenPath      = "/token"
	UserinfoPath   = "/userinfo"
	IntrospectPath = "/introspect"
	RevocationPath = "/revoke"

	VerifyPath  = "/verify"
	RecoverPath = "/recover"
	LogoutPath  = "/logout"

	UserPath = "/user"
)

func (s *server) setupRouting() {
	router := mux.NewRouter().UseEncodedPath()

	router.Use(withConfigMiddleware(s.config))
	router.Use(requestIDMiddleware(s.config.API, s.log))
	router.Use(userIPMiddleware())
	router.Use(loggingContextMiddleware(s.log))

	// Add API router.
	s.registerAPIRouter(router, s.config.API)

	router.NotFoundHandler = serverError(s.log, badRequestError(http.StatusText(http.StatusNotFound)))
	router.MethodNotAllowedHandler = serverError(s.log, httpError(http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed)))

	var h http.Handler
	h = router
	h = handlers.CORS(
		handlers.AllowedMethods([]string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}),
		handlers.AllowedHeaders([]string{xhttp.Accept, xhttp.Authorization, xhttp.ContentType, xhttp.XUseCookie}),
		handlers.AllowCredentials(),
	)(h)
	h = handlers.RecoveryHandler(
		handlers.RecoveryLogger(recoveryHandlerLogger{s.log}),
	)(h)

	s.Handler = h
}

// registerAPIRouter - registers API routes.
func (s *server) registerAPIRouter(router *mux.Router, c *config.APIConfig) {
	// API Router
	apiRouter := router.PathPrefix(SlashSeparator).Subrouter()

	var csrfMiddleware mux.MiddlewareFunc

	if c.CSRF.Enabled {
		csrfCookieName := fmt.Sprintf("%s.csrf-token", authzy.AppName)
		if c.Secure {
			csrfCookieName = "__Host-" + csrfCookieName
		}

		csrfMiddleware = csrf.Protect(
			[]byte(c.CSRF.AuthKey),
			csrf.CookieName(csrfCookieName),
			csrf.ErrorHandler(serverError(s.log, forbiddenError("CSRF token invalid"))),
			csrf.SameSite(csrf.SameSiteLaxMode),
			csrf.Secure(c.Secure),
		)
	}

	var routers []*mux.Router
	routers = append(routers, apiRouter)

	for _, r := range routers {
		// Returns CSRF token in the header for subsequent requests.
		csrfRouter := r.Path(CSRFPath).Subrouter()
		csrfRouter.Methods(http.MethodGet).HandlerFunc(s.CSRFTokenHandler)

		// Creates a new user using their email address.
		signupRouter := r.Path(SignupPath).Subrouter()
		signupRouter.Methods(http.MethodPost).HandlerFunc(s.SignupHandler)

		// Logs in an existing user using their email address.
		// Generates a new JWT.
		r.Path(TokenPath).Methods(http.MethodPost).HandlerFunc(s.TokenHandler)
		// Return a user's profile from the Access Token.
		r.Path(UserinfoPath).Methods(http.MethodGet, http.MethodPost).Handler(
			s.AuthHandler(s.UserinfoHandler),
		)
		// Introspect OAuth2 Tokens.
		r.Path(IntrospectPath).Methods(http.MethodPost).Handler(
			s.AuthHandler(s.IntrospectHandler),
		)
		// Revoke existing tokens.
		r.Path(RevocationPath).Methods(http.MethodPost).HandlerFunc(s.RevocationHandler)

		// Verify exchanges a confirmation or recovery token for a refresh token.
		r.Path(VerifyPath).Methods(http.MethodPost, http.MethodGet).HandlerFunc(s.VerifyHandler)
		// Sends a reset request to an email address.
		r.Path(RecoverPath).Methods(http.MethodPost).HandlerFunc(s.RecoverHandler)

		// Removes a logged-in session.
		r.Path(LogoutPath).Methods(http.MethodPost).Handler(
			s.AuthHandler(s.LogoutHandler),
		)

		// Gets the user details.
		r.Path(UserPath).Methods(http.MethodGet).Handler(
			s.AuthHandler(s.UserGetHandler),
		)
		// Updates the user data.
		r.Path(UserPath).Methods(http.MethodPost).Handler(
			s.AuthHandler(s.UserUpdateHandler),
		)

		if c.CSRF.Enabled {
			csrfRouter.Use(csrfMiddleware)
			signupRouter.Use(csrfMiddleware)
		}
	}
}

func serverError(log logger.Logger, err error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleError(w, r, log, err)
	})
}
