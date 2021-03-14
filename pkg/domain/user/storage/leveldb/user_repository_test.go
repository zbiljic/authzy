package leveldb_test

import (
	"testing"

	database_leveldb "github.com/zbiljic/authzy/pkg/database/leveldb"
	"github.com/zbiljic/authzy/pkg/domain/user"
	"github.com/zbiljic/authzy/pkg/domain/user/storage/leveldb"
	"github.com/zbiljic/authzy/pkg/domain/user/storage/test"
)

func TestLevelDBUserRepository(t *testing.T) {
	test.Run(t, func() func(t *testing.T) (user.UserRepository, func()) {
		return func(t *testing.T) (user.UserRepository, func()) {
			db, cleanup := database_leveldb.Fixture()

			repo, err := leveldb.NewUserRepository(db, test.TestPrefix)
			if err != nil {
				t.Fatal(err)
			}

			return repo, cleanup
		}
	})
}
