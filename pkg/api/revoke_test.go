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
	"github.com/zbiljic/authzy/pkg/domain/user"
)

type RevocationTestSuite struct {
	suite.Suite

	Server *TestServer
}

//nolint:errcheck
func (ts *RevocationTestSuite) SetupTest() {
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

func TestRevocation(t *testing.T) {
	ts := &RevocationTestSuite{}

	ts.Server, _ = newTestServer(t, testServerOptions{})
	defer ts.Server.API.Close()

	suite.Run(t, ts)
}

func (ts *RevocationTestSuite) TestHandler() {
	t := ts.T()

	user, err := ts.Server.UserUsecase.FindUserByEmail(context.Background(), "test@example.com")
	require.NoError(t, err)

	refreshToken, err := ts.Server.RefreshTokenUsecase.GrantAuthenticatedUser(context.Background(), user)
	require.NoError(t, err)

	assert.False(t, refreshToken.Revoked)

	apitest.New().
		Handler(ts.Server.API).
		Post(api.RevocationPath).
		FormData("token", refreshToken.Token).
		Expect(t).
		Status(http.StatusOK).
		End()

	refreshToken, err = ts.Server.RefreshTokenUsecase.FindRefreshTokenByID(context.Background(), refreshToken.ID)
	require.NoError(t, err)

	assert.True(t, refreshToken.Revoked)
}
