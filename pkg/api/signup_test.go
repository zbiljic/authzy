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
	"github.com/zbiljic/authzy/pkg/domain/account"
	"github.com/zbiljic/authzy/pkg/domain/user"
	xhttp "github.com/zbiljic/authzy/pkg/http"
	"github.com/zbiljic/authzy/pkg/ulid"
)

type SignupTestSuite struct {
	suite.Suite

	Server *TestServer
	Config *config.Config
}

//nolint:errcheck
func (ts *SignupTestSuite) SetupTest() {
	// truncate
	ts.Server.AccountRepository.DeleteAll(context.Background())
	ts.Server.RefreshTokenRepository.DeleteAll(context.Background())
	ts.Server.UserRepository.DeleteAll(context.Background())
}

func TestSignup(t *testing.T) {
	ts := &SignupTestSuite{}

	ts.Server, ts.Config = newTestServer(t, testServerOptions{
		Config: &config.Config{
			API: &config.APIConfig{
				CSRF: &config.CSRFConfig{
					AuthKey: "test",
				},
				Mailer: &config.MailerConfig{
					Autoconfirm: true,
				},
			},
		},
	})
	defer ts.Server.API.Close()

	suite.Run(t, ts)
}

// TestSignup tests API `/signup` route.
func (ts *SignupTestSuite) TestSignup() {
	ctx := context.Background()

	t := ts.T()

	t.Run("ok", func(t *testing.T) {
		csrfToken, cookie := csrfTokenHelper(t, ts.Server.API)

		req := &api.SignupRequest{
			Email:    "test@example.com",
			Username: "test",
			Password: "password",
		}

		resp := &api.SignupResponse{}

		apitest.New().
			Handler(ts.Server.API).
			Post(api.SignupPath).
			Header(xhttp.XCSRFToken, csrfToken).
			Cookie(cookie.Name, cookie.Value).
			JSON(req).
			Expect(t).
			Status(http.StatusCreated).
			End().
			JSON(resp)

		assert.NotEmpty(t, resp.ID)
		assert.Equal(t, req.Email, resp.Email)
		assert.Equal(t, req.Username, resp.Username)
		assert.Equal(t, "", resp.Picture)

		user, err := ts.Server.UserUsecase.FindUserByID(ctx, resp.ID)
		require.NoError(t, err)

		assert.NotNil(t, user)
		assert.Equal(t, resp.ID, user.ID)
		assert.WithinDuration(t, time.Now(), *user.ValidSince, 10*time.Second)
		assert.WithinDuration(t, time.Now(), user.CreatedAt, 10*time.Second)
		assert.WithinDuration(t, time.Now(), user.UpdatedAt, 10*time.Second)

		expectedPasswordHash, err := ts.Server.Hasher.Generate(ctx, []byte(req.Password))
		require.NoError(t, err)

		assert.Equal(t, string(expectedPasswordHash), user.PasswordHash)

		accounts, err := ts.Server.AccountUsecase.FindAllForUser(ctx, user.ID)
		require.NoError(t, err)

		assert.Len(t, accounts, 1)
		assert.Equal(t, user.ID, accounts[0].UserID)
		assert.Equal(t, account.ProviderTypePassword, accounts[0].Provider)
		assert.Equal(t, user.Email, accounts[0].FederatedID)
	})

	t.Run("empty request", func(t *testing.T) {
		csrfToken, cookie := csrfTokenHelper(t, ts.Server.API)

		resp := &api.HTTPError{}

		apitest.New().
			Handler(ts.Server.API).
			Post(api.SignupPath).
			Header(xhttp.XCSRFToken, csrfToken).
			Cookie(cookie.Name, cookie.Value).
			Expect(t).
			Status(http.StatusBadRequest).
			End().
			JSON(&resp)

		assert.Equal(t, resp.Message, "Empty request body")
	})

	t.Run("no email", func(t *testing.T) {
		csrfToken, cookie := csrfTokenHelper(t, ts.Server.API)

		req := &api.SignupRequest{}

		resp := &api.HTTPError{}

		apitest.New().
			Handler(ts.Server.API).
			Post(api.SignupPath).
			Header(xhttp.XCSRFToken, csrfToken).
			Cookie(cookie.Name, cookie.Value).
			JSON(req).
			Expect(t).
			Status(http.StatusUnprocessableEntity).
			End().
			JSON(&resp)

		assert.Equal(t, resp.Message, "Account creation requires a valid email")
	})

	t.Run("no password", func(t *testing.T) {
		csrfToken, cookie := csrfTokenHelper(t, ts.Server.API)

		req := &api.SignupRequest{
			Email: "test@example.com",
		}

		resp := &api.HTTPError{}

		apitest.New().
			Handler(ts.Server.API).
			Post(api.SignupPath).
			Header(xhttp.XCSRFToken, csrfToken).
			Cookie(cookie.Name, cookie.Value).
			JSON(req).
			Expect(t).
			Status(http.StatusUnprocessableEntity).
			End().
			JSON(&resp)

		assert.Equal(t, resp.Message, "Account creation requires a valid password")
	})
}

// TestSignupTwice checks to make sure the same email cannot be registered twice.
func (ts *SignupTestSuite) TestSignupTwice() {
	t := ts.T()

	csrfToken, cookie := csrfTokenHelper(t, ts.Server.API)

	req := &api.SignupRequest{
		Email:    "test2@example.com",
		Username: "test2",
		Password: "password",
	}

	resp1 := &api.SignupResponse{}

	apitest.New().
		Handler(ts.Server.API).
		Post(api.SignupPath).
		Header(xhttp.XCSRFToken, csrfToken).
		Cookie(cookie.Name, cookie.Value).
		JSON(req).
		Expect(t).
		Status(http.StatusCreated).
		End().
		JSON(resp1)

	assert.NotEmpty(t, resp1.ID)
	assert.Equal(t, req.Email, resp1.Email)
	assert.Equal(t, req.Username, resp1.Username)

	resp2 := &api.HTTPError{}

	apitest.New().
		Handler(ts.Server.API).
		Post(api.SignupPath).
		Header(xhttp.XCSRFToken, csrfToken).
		Cookie(cookie.Name, cookie.Value).
		JSON(req).
		Expect(t).
		Status(http.StatusBadRequest).
		End().
		JSON(resp2)

	assert.Equal(t, resp2.Message, "A user with this email address has already been registered")
}

func (ts *SignupTestSuite) TestVerifySignup() {
	ctx := context.Background()
	now := time.Now()

	t := ts.T()

	createUserRequest := user.User{
		Email:              "test@example.com",
		Username:           "test",
		Password:           "test",
		ConfirmationToken:  ulid.ZeroULID().String(),
		ConfirmationSentAt: &now,
	}

	_, err := ts.Server.UserUsecase.CreateUser(ctx, &createUserRequest)
	require.NoError(t, err)

	user, err := ts.Server.UserUsecase.FindUserByEmail(ctx, createUserRequest.Email)
	require.NoError(t, err)

	assert.NotEmpty(t, user.ConfirmationToken)
	assert.NotNil(t, user.ConfirmationSentAt)
	assert.False(t, user.EmailVerified)
	assert.Nil(t, user.ValidSince)

	req := &api.VerifyRequest{
		Type:  "signup",
		Token: user.ConfirmationToken,
	}

	resp := &api.AccessTokenResponse{}

	apitest.New().
		Handler(ts.Server.API).
		Post(api.VerifyPath).
		JSON(req).
		Expect(t).
		Status(http.StatusOK).
		End().
		JSON(resp)

	assert.NotEmpty(t, resp.TokenType)
	assert.Equal(t, resp.ExpiresIn, ts.Config.API.JWT.Exp)

	user, err = ts.Server.UserUsecase.FindUserByEmail(ctx, createUserRequest.Email)
	require.NoError(t, err)

	assert.Empty(t, user.ConfirmationToken)
	assert.Nil(t, user.ConfirmationSentAt)
	assert.True(t, user.EmailVerified)
	assert.NotNil(t, user.ValidSince)
	assert.WithinDuration(t, now, *user.ValidSince, 10*time.Second)
}
