package mailer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSiteURL(t *testing.T) {
	cases := []struct {
		ReferrerURL string
		SiteURL     string
		Path        string
		RawQuery    string
		Expected    string
	}{
		{"", "https://test.example.com", "/templates/confirm.html", "", "https://test.example.com/templates/confirm.html"},
		{"", "https://test.example.com/removedpath", "/templates/confirm.html", "", "https://test.example.com/templates/confirm.html"},
		{"", "https://test.example.com/", "/trailingslash/", "", "https://test.example.com/trailingslash/"},
		{"", "https://test.example.com", "f", "query", "https://test.example.com/f?query"},
		{"https://test.example.com/admin", "https://test.example.com", "", "query", "https://test.example.com/admin?query"},
		{"https://test.example.com/admin", "https://test.example.com", "f", "query", "https://test.example.com/f?query"},
		{"", "https://test.example.com", "", "query", "https://test.example.com?query"},
	}

	for _, c := range cases {
		act, err := getSiteURL(c.ReferrerURL, c.SiteURL, c.Path, c.RawQuery)
		assert.NoError(t, err, c.Expected)
		assert.Equal(t, c.Expected, act)
	}
}
