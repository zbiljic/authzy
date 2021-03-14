package api_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/zbiljic/authzy/pkg/api"
	"github.com/zbiljic/authzy/pkg/config"
	"github.com/zbiljic/authzy/pkg/domain/user"
	xhttp "github.com/zbiljic/authzy/pkg/http"
)

type IntrospectTestSuite struct {
	suite.Suite

	Server *TestServer
	Config *config.Config
}

//nolint:errcheck
func (ts *IntrospectTestSuite) SetupTest() {
	// truncate
	ts.Server.AccountRepository.DeleteAll(context.Background())
	ts.Server.RefreshTokenRepository.DeleteAll(context.Background())
	ts.Server.UserRepository.DeleteAll(context.Background())

	// create test user
	createUserRequest := user.User{
		Email:    "test@example.com",
		Username: "test",
		Password: "password",
	}
	user, err := ts.Server.UserUsecase.CreateUser(context.Background(), &createUserRequest)
	require.NoError(ts.T(), err)

	_, err = ts.Server.UserUsecase.ConfirmUser(context.Background(), user.ID)
	require.NoError(ts.T(), err)
}

func TestIntrospect(t *testing.T) {
	ts := &IntrospectTestSuite{}

	ts.Server, ts.Config = newTestServer(t, testServerOptions{})
	defer ts.Server.API.Close()

	suite.Run(t, ts)
}

func (ts *IntrospectTestSuite) TestAccessToken() {
	t := ts.T()

	t.Run("token", func(t *testing.T) {
		auth := authTokenHelper(t, ts.Server.API, "test@example.com", "password")

		resp := &api.Introspection{}

		apitest.New().
			Handler(ts.Server.API).
			Post(api.IntrospectPath).
			Header(xhttp.Authorization, fmt.Sprintf("%s %s", auth.TokenType, auth.Token)).
			FormData("token", auth.Token).
			Expect(t).
			Status(http.StatusOK).
			End().
			JSON(resp)

		assert.NotNil(t, resp)

		assert.True(t, resp.Active)
		assert.Equal(t, auth.TokenType, resp.TokenType)
		assert.Equal(t, "access_token", resp.TokenUse)

		user, err := ts.Server.UserUsecase.FindUserByEmail(context.Background(), "test@example.com")
		require.NoError(t, err)

		assert.Equal(t, user.ID, resp.Subject)
		assert.WithinDuration(t, time.Now().Add(time.Duration(ts.Config.API.JWT.Exp)*time.Second), time.Unix(resp.ExpiresAt, 0), 10*time.Second)
		assert.WithinDuration(t, time.Now(), time.Unix(resp.IssuedAt, 0), 10*time.Second)
		assert.WithinDuration(t, time.Now(), time.Unix(resp.NotBefore, 0), 10*time.Second)
		assert.Equal(t, user.Username, resp.Username)
		assert.Contains(t, resp.Extra, "email")
		assert.Equal(t, user.Email, resp.Extra["email"])
	})

	t.Run("type_hint", func(t *testing.T) {
		auth := authTokenHelper(t, ts.Server.API, "test@example.com", "password")

		resp := &api.Introspection{}

		apitest.New().
			Handler(ts.Server.API).
			Post(api.IntrospectPath).
			Header(xhttp.Authorization, fmt.Sprintf("%s %s", auth.TokenType, auth.Token)).
			FormData("token", auth.Token).
			FormData("token_type_hint", "access_token").
			Expect(t).
			Status(http.StatusOK).
			End().
			JSON(resp)

		assert.NotNil(t, resp)

		assert.True(t, resp.Active)
		assert.Equal(t, auth.TokenType, resp.TokenType)
		assert.Equal(t, "access_token", resp.TokenUse)

		user, err := ts.Server.UserUsecase.FindUserByEmail(context.Background(), "test@example.com")
		require.NoError(t, err)

		assert.Equal(t, user.ID, resp.Subject)
		assert.WithinDuration(t, time.Now().Add(time.Duration(ts.Config.API.JWT.Exp)*time.Second), time.Unix(resp.ExpiresAt, 0), 10*time.Second)
		assert.WithinDuration(t, time.Now(), time.Unix(resp.IssuedAt, 0), 10*time.Second)
		assert.WithinDuration(t, time.Now(), time.Unix(resp.NotBefore, 0), 10*time.Second)
		assert.Equal(t, user.Username, resp.Username)
		assert.Contains(t, resp.Extra, "email")
		assert.Equal(t, user.Email, resp.Extra["email"])
	})
}

func (ts *IntrospectTestSuite) TestRefreshToken() {
	t := ts.T()

	t.Run("token", func(t *testing.T) {
		auth := authTokenHelper(t, ts.Server.API, "test@example.com", "password")

		resp := &api.Introspection{}

		apitest.New().
			Handler(ts.Server.API).
			Post(api.IntrospectPath).
			Header(xhttp.Authorization, fmt.Sprintf("%s %s", auth.TokenType, auth.Token)).
			FormData("token", auth.RefreshToken).
			Expect(t).
			Status(http.StatusOK).
			End().
			JSON(resp)

		assert.NotNil(t, resp)

		assert.True(t, resp.Active)
		assert.Equal(t, auth.TokenType, resp.TokenType)
		assert.Equal(t, "refresh_token", resp.TokenUse)

		user, err := ts.Server.UserUsecase.FindUserByEmail(context.Background(), "test@example.com")
		require.NoError(t, err)

		assert.Equal(t, user.ID, resp.Subject)
		assert.WithinDuration(t, time.Now(), time.Unix(resp.IssuedAt, 0), 10*time.Second)
		assert.WithinDuration(t, time.Now(), time.Unix(resp.NotBefore, 0), 10*time.Second)
		assert.Equal(t, user.Username, resp.Username)
		assert.Contains(t, resp.Extra, "email")
		assert.Equal(t, user.Email, resp.Extra["email"])
	})

	t.Run("type_hint", func(t *testing.T) {
		auth := authTokenHelper(t, ts.Server.API, "test@example.com", "password")

		resp := &api.Introspection{}

		apitest.New().
			Handler(ts.Server.API).
			Post(api.IntrospectPath).
			Header(xhttp.Authorization, fmt.Sprintf("%s %s", auth.TokenType, auth.Token)).
			FormData("token", auth.RefreshToken).
			FormData("token_type_hint", "refresh_token").
			Expect(t).
			Status(http.StatusOK).
			End().
			JSON(resp)

		assert.NotNil(t, resp)

		assert.True(t, resp.Active)
		assert.Equal(t, auth.TokenType, resp.TokenType)
		assert.Equal(t, "refresh_token", resp.TokenUse)

		user, err := ts.Server.UserUsecase.FindUserByEmail(context.Background(), "test@example.com")
		require.NoError(t, err)

		assert.Equal(t, user.ID, resp.Subject)
		assert.WithinDuration(t, time.Now(), time.Unix(resp.IssuedAt, 0), 10*time.Second)
		assert.WithinDuration(t, time.Now(), time.Unix(resp.NotBefore, 0), 10*time.Second)
		assert.Equal(t, user.Username, resp.Username)
		assert.Contains(t, resp.Extra, "email")
		assert.Equal(t, user.Email, resp.Extra["email"])
	})
}
