package leveldb_test

import (
	"testing"

	database_leveldb "github.com/zbiljic/authzy/pkg/database/leveldb"
	"github.com/zbiljic/authzy/pkg/domain/account"
	"github.com/zbiljic/authzy/pkg/domain/account/storage/leveldb"
	"github.com/zbiljic/authzy/pkg/domain/account/storage/test"
)

func TestLevelDBUserRepository(t *testing.T) {
	test.Run(t, func() func(t *testing.T) (account.AccountRepository, func()) {
		return func(t *testing.T) (account.AccountRepository, func()) {
			db, cleanup := database_leveldb.Fixture()

			repo, err := leveldb.NewAccountRepository(db, test.TestPrefix)
			if err != nil {
				t.Fatal(err)
			}

			return repo, cleanup
		}
	})
}
