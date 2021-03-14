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
)

type RecoverTestSuite struct {
	suite.Suite

	Server *TestServer
	Config *config.Config
}

//nolint:errcheck
func (ts *RecoverTestSuite) SetupTest() {
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

func TestRecover(t *testing.T) {
	ts := &RecoverTestSuite{}

	ts.Server, ts.Config = newTestServer(t, testServerOptions{})
	defer ts.Server.API.Close()

	suite.Run(t, ts)
}

func (ts *RecoverTestSuite) TestFirstRecovery() {
	t := ts.T()

	user, err := ts.Server.UserUsecase.FindUserByEmail(context.Background(), "test@example.com")
	require.NoError(t, err)

	user.RecoverySentAt = &time.Time{}

	_, err = ts.Server.UserUsecase.UpdateUser(context.Background(), user)
	require.NoError(t, err)

	req := &api.RecoverRequest{
		Email: "test@example.com",
	}

	apitest.New().
		Handler(ts.Server.API).
		Post(api.RecoverPath).
		JSON(req).
		Expect(t).
		Status(http.StatusOK).
		End()

	user, err = ts.Server.UserUsecase.FindUserByEmail(context.Background(), "test@example.com")
	require.NoError(t, err)

	assert.WithinDuration(t, time.Now(), *user.RecoverySentAt, 10*time.Second)
}

func (ts *RecoverTestSuite) TestNoEmailSent() {
	t := ts.T()

	user, err := ts.Server.UserUsecase.FindUserByEmail(context.Background(), "test@example.com")
	require.NoError(t, err)

	// half of max frequency configuration
	recoveryTime := time.Now().UTC().Add(-1 * (ts.Config.SMTP.MaxFrequency / 2))
	user.RecoverySentAt = &recoveryTime

	_, err = ts.Server.UserUsecase.UpdateUser(context.Background(), user)
	require.NoError(t, err)

	req := &api.RecoverRequest{
		Email: "test@example.com",
	}

	apitest.New().
		Handler(ts.Server.API).
		Post(api.RecoverPath).
		JSON(req).
		Expect(t).
		Status(http.StatusTooManyRequests).
		End()

	user, err = ts.Server.UserUsecase.FindUserByEmail(context.Background(), "test@example.com")
	require.NoError(t, err)

	// ensure it did not send a new email
	u1 := recoveryTime.Round(time.Second).Unix()
	u2 := user.RecoverySentAt.Round(time.Second).Unix()
	assert.Equal(t, u1, u2)
}

func (ts *RecoverTestSuite) TestNewEmailSent() {
	t := ts.T()

	user, err := ts.Server.UserUsecase.FindUserByEmail(context.Background(), "test@example.com")
	require.NoError(t, err)

	recoveryTime := time.Now().UTC().Add(-20 * time.Minute)
	user.RecoverySentAt = &recoveryTime

	_, err = ts.Server.UserUsecase.UpdateUser(context.Background(), user)
	require.NoError(t, err)

	req := &api.RecoverRequest{
		Email: "test@example.com",
	}

	apitest.New().
		Handler(ts.Server.API).
		Post(api.RecoverPath).
		JSON(req).
		Expect(t).
		Status(http.StatusOK).
		End()

	user, err = ts.Server.UserUsecase.FindUserByEmail(context.Background(), "test@example.com")
	require.NoError(t, err)

	// ensure it sent a new email
	assert.WithinDuration(t, time.Now(), *user.RecoverySentAt, 10*time.Second)
}
