package api_test

import (
	"crypto/rand"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/lestrrat-go/jwx/jwk"

	"github.com/zbiljic/authzy/pkg/api"
	"github.com/zbiljic/authzy/pkg/config"
	"github.com/zbiljic/authzy/pkg/domain/account"
	account_jsonmutexdb "github.com/zbiljic/authzy/pkg/domain/account/storage/jsonmutexdb"
	accountuc "github.com/zbiljic/authzy/pkg/domain/account/usecases"
	"github.com/zbiljic/authzy/pkg/domain/refreshtoken"
	refreshtoken_jsonmutexdb "github.com/zbiljic/authzy/pkg/domain/refreshtoken/storage/jsonmutexdb"
	refreshtokenuc "github.com/zbiljic/authzy/pkg/domain/refreshtoken/usecases"
	"github.com/zbiljic/authzy/pkg/domain/user"
	user_jsonmutexdb "github.com/zbiljic/authzy/pkg/domain/user/storage/jsonmutexdb"
	useruc "github.com/zbiljic/authzy/pkg/domain/user/usecases"
	"github.com/zbiljic/authzy/pkg/hash"
	mockhasher "github.com/zbiljic/authzy/pkg/hash/mock"
	"github.com/zbiljic/authzy/pkg/jwt"
	"github.com/zbiljic/authzy/pkg/logger"
	"github.com/zbiljic/authzy/pkg/testutil"
)

type TestServer struct {
	Addr string
	API  api.Service

	Hasher                 hash.Hasher
	AccountRepository      account.AccountRepository
	AccountUsecase         account.AccountUsecase
	RefreshTokenRepository refreshtoken.RefreshTokenRepository
	RefreshTokenUsecase    refreshtoken.RefreshTokenUsecase
	UserRepository         user.UserRepository
	UserUsecase            user.UserUsecase
}

type testServerOptions struct {
	Log                    logger.Logger
	Config                 *config.Config
	Hasher                 hash.Hasher
	AccountRepository      account.AccountRepository
	RefreshTokenRepository refreshtoken.RefreshTokenRepository
	UserRepository         user.UserRepository
	JwtService             jwt.Service
}

func newTestServer(t *testing.T, o testServerOptions) (*TestServer, *config.Config) {
	t.Helper()

	if o.Log == nil {
		log := testutil.Logger{}
		log.SetOutput(io.Discard)
		o.Log = log
	}
	if o.Config == nil {
		conf, _ := config.LoadConfig("")
		o.Config = conf
	} else {
		conf, _ := config.LoadConfig("")
		conf, _ = conf.Merge(o.Config)
		o.Config = conf
	}
	if o.Hasher == nil {
		o.Hasher = mockhasher.NewMockHasher()
	}
	if o.AccountRepository == nil {
		o.AccountRepository, _ = account_jsonmutexdb.NewAccountRepository(nil, "")
	}
	if o.RefreshTokenRepository == nil {
		o.RefreshTokenRepository, _ = refreshtoken_jsonmutexdb.NewRefreshTokenRepository(nil, "")
	}
	if o.UserRepository == nil {
		o.UserRepository, _ = user_jsonmutexdb.NewUserRepository(nil, "")
	}
	if o.JwtService == nil {
		if o.Config.API.JWT.ClaimsNamespace == "" {
			o.Config.API.JWT.ClaimsNamespace = "https://example.test/jwt/claims"
		}

		if o.Config.API.JWT.KeysJSON == "" {
			// generate JWT key set
			keySet := jwk.NewSet()

			sharedKey := make([]byte, 64)
			rand.Read(sharedKey) //nolint:errcheck

			keyID := "test"
			key, err := jwk.New(sharedKey)
			if err != nil {
				t.Fatal(err)
			}

			err = key.Set(jwk.KeyIDKey, keyID)
			if err != nil {
				t.Fatal(err)
			}

			keySet.Add(key)

			buf, _ := json.Marshal(keySet)

			o.Config.API.JWT.DefaultKey = keyID
			o.Config.API.JWT.KeysJSON = string(buf)
		}

		jwtService, err := jwt.NewService(o.Config.API.JWT)
		if err != nil {
			t.Fatal(err)
		}

		o.JwtService = jwtService
	}

	accountUsecase := accountuc.NewAccountUsecase(o.AccountRepository)
	refreshTokenUsecase := refreshtokenuc.NewRefreshTokenUsecase(o.RefreshTokenRepository)
	userUsecase := useruc.NewUserUsecase(o.Hasher, o.UserRepository)

	s := api.New(
		o.Log,
		o.Config,
		o.JwtService,
		accountUsecase,
		refreshTokenUsecase,
		userUsecase,
	)

	ts := httptest.NewServer(s)
	t.Cleanup(ts.Close)

	testServer := &TestServer{
		Addr:                   ts.Listener.Addr().String(),
		API:                    s,
		Hasher:                 o.Hasher,
		AccountRepository:      o.AccountRepository,
		AccountUsecase:         accountUsecase,
		RefreshTokenRepository: o.RefreshTokenRepository,
		RefreshTokenUsecase:    refreshTokenUsecase,
		UserRepository:         o.UserRepository,
		UserUsecase:            userUsecase,
	}

	return testServer, o.Config
}
