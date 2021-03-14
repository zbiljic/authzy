package api

import (
	"net/http"

	"github.com/gorilla/csrf"

	xhttp "github.com/zbiljic/authzy/pkg/http"
)

func (s *server) CSRFTokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(xhttp.XCSRFToken, csrf.Token(r))
}
