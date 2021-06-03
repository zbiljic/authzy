package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	"github.com/zbiljic/authzy/pkg/domain/user"
)

var (
	// ContentTypeJSONHeader is default value of if "Content-Type" header
	// in HTTP response.
	ContentTypeJSONHeader = "application/json; charset=utf-8"
	// EscapeHTML specifies whether problematic HTML characters
	// should be escaped inside JSON quoted strings.
	EscapeHTML = false
)

func mustSendJSON(w http.ResponseWriter, status int, response interface{}) {
	if err := sendJSON(w, status, response); err != nil {
		panic(fmt.Sprintf("Unable to send JSON: %v", err))
	}
}

func sendJSON(w http.ResponseWriter, status int, response interface{}) error {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(EscapeHTML)
	if err := enc.Encode(response); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error encoding json response: %v", response))
	}
	w.WriteHeader(status)
	w.Header().Set("Content-Type", ContentTypeJSONHeader)
	_, err := w.Write(b.Bytes())
	return err
}

func (s *server) getReferrer(r *http.Request) string {
	ctx := r.Context()
	config := getConfig(ctx)

	referrer := ""
	if reqref := r.Referer(); reqref != "" {
		base, berr := url.Parse(config.SiteURL)
		refurl, rerr := url.Parse(reqref)
		// As long as the referrer came from the site, we will redirect back there
		if berr == nil && rerr == nil && base.Hostname() == refurl.Hostname() {
			referrer = reqref
		}
	}
	return referrer
}

// validateRedirectURL ensures any redirect URL is from a safe origin.
func (s *server) validateRedirectURL(r *http.Request, reqref string) string {
	ctx := r.Context()
	config := getConfig(ctx)

	redirectURL := config.SiteURL
	if reqref != "" {
		base, berr := url.Parse(config.SiteURL)
		refurl, rerr := url.Parse(reqref)
		// As long as the referrer came from the site, we will redirect back there
		if berr == nil && rerr == nil && base.Hostname() == refurl.Hostname() {
			redirectURL = reqref
		}
	}
	return redirectURL
}

func (s *server) getUserFromToken(ctx context.Context) (*user.User, error) {
	jwtToken := *getToken(ctx)

	sub := jwtToken.Subject()
	if sub == "" {
		return nil, errors.New("Could not read 'sub' claim")
	}

	return s.userUsecase.FindUserByID(ctx, sub)
}
