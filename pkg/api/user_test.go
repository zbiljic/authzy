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

type UserTestSuite struct {
	suite.Suite

	Server *TestServer
}

//nolint:errcheck
func (ts *UserTestSuite) SetupTest() {
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

func TestUser(t *testing.T) {
	ts := &UserTestSuite{}

	ts.Server, _ = newTestServer(t, testServerOptions{})
	defer ts.Server.API.Close()

	suite.Run(t, ts)
}

func (ts *UserTestSuite) TestUpdatePassword() {
	t := ts.T()

	auth := authTokenHelper(t, ts.Server.API, "test@example.com", "password")

	req := &api.UserUpdateRequest{
		Password: "new_password",
	}

	resp := &api.UserResponse{}

	apitest.New().
		Handler(ts.Server.API).
		Post(api.UserPath).
		Header(xhttp.Authorization, fmt.Sprintf("%s %s", auth.TokenType, auth.Token)).
		JSON(req).
		Expect(t).
		Status(http.StatusOK).
		End().
		JSON(resp)

	assert.Equal(t, "test@example.com", resp.Email)

	user, err := ts.Server.UserUsecase.FindUserByEmail(context.Background(), "test@example.com")
	require.NoError(t, err)

	expectedPasswordHash, err := ts.Server.Hasher.Generate(context.Background(), []byte("new_password"))
	require.NoError(t, err)

	assert.Equal(t, string(expectedPasswordHash), user.PasswordHash)
}
