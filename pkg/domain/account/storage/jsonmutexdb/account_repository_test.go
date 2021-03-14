package jsonmutexdb_test

import (
	"testing"

	"github.com/zbiljic/authzy/pkg/domain/account"
	"github.com/zbiljic/authzy/pkg/domain/account/storage/jsonmutexdb"
	"github.com/zbiljic/authzy/pkg/domain/account/storage/test"
)

func TestJSONMutexDBRefreshTokenRepository(t *testing.T) {
	test.Run(t, func() func(t *testing.T) (account.AccountRepository, func()) {
		return func(t *testing.T) (account.AccountRepository, func()) {
			repo, err := jsonmutexdb.NewAccountRepository(nil, test.TestPrefix)
			if err != nil {
				t.Fatal(err)
			}

			return repo, func() {}
		}
	})
}
