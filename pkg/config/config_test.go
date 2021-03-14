package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zbiljic/authzy/pkg/config"
	xhttp "github.com/zbiljic/authzy/pkg/http"
)

func TestMain(m *testing.M) {
	exitCode := m.Run()
	os.Clearenv()
	os.Exit(exitCode)
}

func TestRequired(t *testing.T) {
	os.Setenv("AUTHZY_DATABASE_TYPE", "jsonmutexdb")
	os.Setenv("AUTHZY_DATABASE_JSONMUTEXDB_DATA_DIR", "fake")
	os.Setenv("AUTHZY_API_CSRF_AUTH_KEY", "32-byte-long-auth-key------------")
	os.Setenv("AUTHZY_API_JWT_CLAIMS_NAMESPACE", "https://example.test/jwt/claims")
	os.Setenv("AUTHZY_API_JWT_DEFAULT_KEY", "test")
	os.Setenv("AUTHZY_API_JWT_KEYS", "{}")

	conf, err := config.LoadConfig("")
	require.NoError(t, err)

	require.NotNil(t, conf)
	assert.Equal(t, "jsonmutexdb", conf.Database.Type)
}

func TestDefaults(t *testing.T) {
	conf, _ := config.LoadConfig("")

	require.NotNil(t, conf)
	assert.Equal(t, xhttp.XRequestID, conf.API.RequestIDHeader)
}
