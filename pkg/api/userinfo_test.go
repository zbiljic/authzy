package api_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/zbiljic/authzy/pkg/api"
	"github.com/zbiljic/authzy/pkg/domain/user"
	xhttp "github.com/zbiljic/authzy/pkg/http"
)

type UserinfoTestSuite struct {
	suite.Suite

	Server *TestServer
}

//nolint:errcheck
func (ts *UserinfoTestSuite) SetupTest() {
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

func TestUserinfo(t *testing.T) {
	ts := &UserinfoTestSuite{}

	ts.Server, _ = newTestServer(t, testServerOptions{})
	defer ts.Server.API.Close()

	suite.Run(t, ts)
}

func (ts *UserinfoTestSuite) TestHandler() {
	t := ts.T()

	auth := authTokenHelper(t, ts.Server.API, "test@example.com", "password")

	resp := &api.UserinfoResponse{}

	apitest.New().
		Handler(ts.Server.API).
		Get(api.UserinfoPath).
		Header(xhttp.Authorization, fmt.Sprintf("%s %s", auth.TokenType, auth.Token)).
		Expect(t).
		Status(http.StatusOK).
		End().
		JSON(resp)

	assert.NotNil(t, resp)

	user, err := ts.Server.UserUsecase.FindUserByEmail(context.Background(), "test@example.com")
	require.NoError(t, err)

	assert.Equal(t, user.ID, resp.Sub)
	assert.Equal(t, user.Name, resp.Name)
	assert.Equal(t, user.GivenName, resp.GivenName)
	assert.Equal(t, user.FamilyName, resp.FamilyName)
	assert.Equal(t, user.Nickname, resp.Nickname)
	assert.Equal(t, user.Username, resp.PreferredUsername)
	assert.Equal(t, user.Picture, resp.Picture)
	assert.Equal(t, user.Email, resp.Email)
	assert.Equal(t, user.EmailVerified, resp.EmailVerified)
	assert.Equal(t, user.UpdatedAt.Unix(), resp.UpdatedAt)
}
