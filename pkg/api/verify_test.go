package api_test

import (
	"context"
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
	"github.com/zbiljic/authzy/pkg/ulid"
)

type VerifyTestSuite struct {
	suite.Suite

	Server *TestServer
	Config *config.Config
}

//nolint:errcheck
func (ts *VerifyTestSuite) SetupTest() {
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
	_, err := ts.Server.UserUsecase.CreateUser(context.Background(), &createUserRequest)
	require.NoError(ts.T(), err)
}

func TestVerify(t *testing.T) {
	ts := &VerifyTestSuite{}

	ts.Server, ts.Config = newTestServer(t, testServerOptions{})
	defer ts.Server.API.Close()

	suite.Run(t, ts)
}

func (ts *VerifyTestSuite) TestPasswordRecovery() {
	t := ts.T()

	user, err := ts.Server.UserUsecase.FindUserByEmail(context.Background(), "test@example.com")
	require.NoError(t, err)

	user.RecoverySentAt = &time.Time{}

	_, err = ts.Server.UserUsecase.UpdateUser(context.Background(), user)
	require.NoError(t, err)

	reqRecover := &api.RecoverRequest{
		Email: "test@example.com",
	}

	apitest.New().
		Handler(ts.Server.API).
		Post(api.RecoverPath).
		JSON(reqRecover).
		Expect(t).
		Status(http.StatusOK).
		End()

	user, err = ts.Server.UserUsecase.FindUserByEmail(context.Background(), "test@example.com")
	require.NoError(t, err)

	assert.WithinDuration(t, time.Now(), *user.RecoverySentAt, 10*time.Second)

	reqVerify := &api.VerifyRequest{
		Type:  "recovery",
		Token: user.RecoveryToken,
	}

	respVerify := &api.AccessTokenResponse{}

	apitest.New().
		Handler(ts.Server.API).
		Post(api.VerifyPath).
		JSON(reqVerify).
		Expect(t).
		Status(http.StatusOK).
		End().
		JSON(respVerify)

	assert.NotEmpty(t, respVerify.TokenType)
	assert.Equal(t, respVerify.ExpiresIn, ts.Config.API.JWT.Exp)

	user, err = ts.Server.UserUsecase.FindUserByEmail(context.Background(), "test@example.com")
	require.NoError(t, err)

	assert.Empty(t, user.RecoveryToken)
	assert.Nil(t, user.RecoverySentAt)
}

func (ts *VerifyTestSuite) TestExpiredConfirmationToken() {
	t := ts.T()

	user, err := ts.Server.UserUsecase.FindUserByEmail(context.Background(), "test@example.com")
	require.NoError(t, err)

	user.ConfirmationToken = ulid.ZeroULID().String()
	sentTime := time.Now().Add(-48 * time.Hour)
	user.ConfirmationSentAt = &sentTime

	user, err = ts.Server.UserUsecase.UpdateUser(context.Background(), user)
	require.NoError(t, err)

	req := &api.VerifyRequest{
		Type:  "signup",
		Token: user.ConfirmationToken,
	}

	resp := &api.HTTPError{}

	apitest.New().
		Handler(ts.Server.API).
		Post(api.VerifyPath).
		JSON(req).
		Expect(t).
		Status(http.StatusGone).
		End().
		JSON(resp)

	assert.Equal(t, resp.Message, "Confirmation token expired")
}

func (ts *VerifyTestSuite) TestExpiredRecoveryToken() {
	t := ts.T()

	user, err := ts.Server.UserUsecase.FindUserByEmail(context.Background(), "test@example.com")
	require.NoError(t, err)

	user.RecoveryToken = ulid.ZeroULID().String()
	sentTime := time.Now().Add(-48 * time.Hour)
	user.RecoverySentAt = &sentTime

	user, err = ts.Server.UserUsecase.UpdateUser(context.Background(), user)
	require.NoError(t, err)

	req := &api.VerifyRequest{
		Type:  "recovery",
		Token: user.RecoveryToken,
	}

	resp := &api.HTTPError{}

	apitest.New().
		Handler(ts.Server.API).
		Post(api.VerifyPath).
		JSON(req).
		Expect(t).
		Status(http.StatusGone).
		End().
		JSON(resp)

	assert.Equal(t, resp.Message, "Recovery token expired")
}
