package api_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/assert"

	"github.com/zbiljic/authzy/pkg/api"
	xhttp "github.com/zbiljic/authzy/pkg/http"
)

func csrfTokenHelper(t *testing.T, r http.Handler) (string, *http.Cookie) {
	t.Helper()

	result := apitest.New().
		Handler(r).
		Get(api.CSRFPath).
		Expect(t).
		End()

	csrfTokenHeaderValue := result.Response.Header.Get(xhttp.XCSRFToken)
	cookieHeaderValue := result.Response.Header.Get(xhttp.SetCookie)

	header := http.Header{}
	header.Add(xhttp.Cookie, cookieHeaderValue)
	request := http.Request{Header: header}

	var csrfCookie *http.Cookie
	cookies := request.Cookies()
	for _, c := range cookies {
		if strings.Contains(c.Name, "csrf") {
			csrfCookie = c

			break
		}
	}

	return csrfTokenHeaderValue, csrfCookie
}

func TestCSRFToken(t *testing.T) {
	server, _ := newTestServer(t, testServerOptions{})

	t.Run("ok", func(t *testing.T) {
		result := apitest.New().
			Handler(server.API).
			Get(api.CSRFPath).
			Expect(t).
			Status(http.StatusOK).
			End()

		csrfTokenHeaderValue := result.Response.Header.Get(xhttp.XCSRFToken)

		assert.NotEmpty(t, csrfTokenHeaderValue)

		cookieHeaderValue := result.Response.Header.Get(xhttp.SetCookie)

		assert.NotEmpty(t, cookieHeaderValue)
	})

	t.Run("helper", func(t *testing.T) {
		csrfToken, cookie := csrfTokenHelper(t, server.API)

		assert.NotEmpty(t, csrfToken)
		assert.NotEmpty(t, cookie)
	})
}
