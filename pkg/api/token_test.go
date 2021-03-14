package api_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/zbiljic/authzy/pkg/api"
	"github.com/zbiljic/authzy/pkg/config"
	"github.com/zbiljic/authzy/pkg/domain/user"
)

func authTokenHelper(t *testing.T, r http.Handler, username, password string) *api.AccessTokenResponse {
	t.Helper()

	resp := &api.AccessTokenResponse{}

	apitest.New().
		Handler(r).
		Post(api.TokenPath).
		FormData("grant_type", "password").
		FormData("username", username).
		FormData("password", password).
		Expect(t).
		Status(http.StatusOK).
		End().
		JSON(resp)

	assert.NotEmpty(t, resp.Token)
	assert.NotEmpty(t, resp.TokenType)

	return resp
}

type TokenTestSuite struct {
	suite.Suite

	Server *TestServer
	Config *config.Config
}

//nolint:errcheck
func (ts *TokenTestSuite) SetupTest() {
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

func TestToken(t *testing.T) {
	ts := &TokenTestSuite{}

	ts.Server, ts.Config = newTestServer(t, testServerOptions{})
	defer ts.Server.API.Close()

	suite.Run(t, ts)
}

func (ts *TokenTestSuite) TestHandler() {
	t := ts.T()

	resp := &api.AccessTokenResponse{}

	apitest.New().
		Handler(ts.Server.API).
		Post(api.TokenPath).
		FormData("grant_type", "password").
		FormData("username", "test@example.com").
		FormData("password", "password").
		Expect(t).
		Status(http.StatusOK).
		End().
		JSON(resp)

	assert.NotEmpty(t, resp.Token)
	assert.NotEmpty(t, resp.TokenType)
	assert.Equal(t, ts.Config.API.JWT.Exp, resp.ExpiresIn)
	assert.NotEmpty(t, resp.RefreshToken)
}
