package api

import (
	"fmt"
	"net/http"

	"github.com/zbiljic/authzy/pkg/logger"
)

type ErrorCause interface {
	Cause() error
}

func handleError(w http.ResponseWriter, r *http.Request, log logger.Logger, err error) {
	ctx := r.Context()
	errorID := getRequestID(ctx)
	switch e := err.(type) {
	case *HTTPError:
		if e.Code >= http.StatusInternalServerError {
			e.ErrorID = errorID
			// this will get us the stack trace too
			log.WithContext(ctx).Error(e.Error())
		} else {
			log.WithContext(ctx).Info(e.Error())
		}
		if jsonErr := sendJSON(w, e.Code, e); jsonErr != nil {
			handleError(w, r, log, jsonErr)
		}
	case *OAuthError:
		log.WithContext(ctx).Info(e.Error())
		if jsonErr := sendJSON(w, http.StatusBadRequest, e); jsonErr != nil {
			handleError(w, r, log, jsonErr)
		}
	case ErrorCause:
		handleError(w, r, log, e.Cause())
	default:
		log.WithContext(ctx).Errorf("Unhandled server error: %s", e.Error(), err)
		// hide real error details from response to prevent info leaks
		w.WriteHeader(http.StatusInternalServerError)
		if _, writeErr := w.Write([]byte(`{"code":500,"msg":"Internal server error","error_id":"` + errorID + `"}`)); writeErr != nil {
			log.WithContext(ctx).Error("Error writing generic error message", writeErr)
		}
	}
}

//
// HTTP
//

func badRequestError(fmtString string, args ...interface{}) *HTTPError {
	return httpError(http.StatusBadRequest, fmtString, args...)
}

func unauthorizedError(fmtString string, args ...interface{}) *HTTPError {
	return httpError(http.StatusUnauthorized, fmtString, args...)
}

func forbiddenError(fmtString string, args ...interface{}) *HTTPError {
	return httpError(http.StatusForbidden, fmtString, args...)
}

func notFoundError(fmtString string, args ...interface{}) *HTTPError {
	return httpError(http.StatusNotFound, fmtString, args...)
}

func goneError(fmtString string, args ...interface{}) *HTTPError {
	return httpError(http.StatusGone, fmtString, args...)
}

func unprocessableEntityError(fmtString string, args ...interface{}) *HTTPError {
	return httpError(http.StatusUnprocessableEntity, fmtString, args...)
}

func tooManyRequestsError(fmtString string, args ...interface{}) *HTTPError {
	return httpError(http.StatusTooManyRequests, fmtString, args...)
}

func internalServerError(fmtString string, args ...interface{}) *HTTPError {
	return httpError(http.StatusInternalServerError, fmtString, args...)
}

// HTTPError is an error with a message and an HTTP status code.
type HTTPError struct {
	Code            int    `json:"code"`
	Message         string `json:"message"`
	InternalError   error  `json:"-"`
	InternalMessage string `json:"-"`
	ErrorID         string `json:"error_id,omitempty"`
}

func (e *HTTPError) Error() string {
	if e.InternalMessage != "" {
		return e.InternalMessage
	}
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

// Cause returns the root cause error.
func (e *HTTPError) Cause() error {
	if e.InternalError != nil {
		return e.InternalError
	}
	return e
}

// WithInternalError adds internal error information to the error.
func (e *HTTPError) WithInternalError(err error) *HTTPError {
	e.InternalError = err
	return e
}

// WithInternalMessage adds internal message information to the error.
func (e *HTTPError) WithInternalMessage(fmtString string, args ...interface{}) *HTTPError {
	e.InternalMessage = fmt.Sprintf(fmtString, args...)
	return e
}

func httpError(code int, fmtString string, args ...interface{}) *HTTPError {
	return &HTTPError{
		Code:    code,
		Message: fmt.Sprintf(fmtString, args...),
	}
}

//
// OAuth 2.0
//

func oauthError(err string, description string) *OAuthError {
	return &OAuthError{Err: err, Description: description}
}

var oauthErrorMap = map[int]string{
	http.StatusBadRequest:          "invalid_request",
	http.StatusUnauthorized:        "unauthorized_client",
	http.StatusForbidden:           "access_denied",
	http.StatusInternalServerError: "server_error",
	http.StatusServiceUnavailable:  "temporarily_unavailable",
}

// OAuthError is the JSON handler for OAuth2 error responses.
type OAuthError struct {
	Err             string `json:"error"`
	Description     string `json:"error_description,omitempty"`
	InternalError   error  `json:"-"`
	InternalMessage string `json:"-"`
}

func (e *OAuthError) Error() string {
	if e.InternalMessage != "" {
		return e.InternalMessage
	}
	return fmt.Sprintf("%s: %s", e.Err, e.Description)
}

// WithInternalError adds internal error information to the error.
func (e *OAuthError) WithInternalError(err error) *OAuthError {
	e.InternalError = err
	return e
}

// WithInternalMessage adds internal message information to the error.
func (e *OAuthError) WithInternalMessage(fmtString string, args ...interface{}) *OAuthError {
	e.InternalMessage = fmt.Sprintf(fmtString, args...)
	return e
}

// Cause returns the root cause error.
func (e *OAuthError) Cause() error {
	if e.InternalError != nil {
		return e.InternalError
	}
	return e
}
