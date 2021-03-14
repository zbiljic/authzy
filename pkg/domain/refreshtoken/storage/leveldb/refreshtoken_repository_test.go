package leveldb_test

import (
	"testing"

	database_leveldb "github.com/zbiljic/authzy/pkg/database/leveldb"
	"github.com/zbiljic/authzy/pkg/domain/refreshtoken"
	"github.com/zbiljic/authzy/pkg/domain/refreshtoken/storage/leveldb"
	"github.com/zbiljic/authzy/pkg/domain/refreshtoken/storage/test"
)

func TestLevelDBUserRepository(t *testing.T) {
	test.Run(t, func() func(t *testing.T) (refreshtoken.RefreshTokenRepository, func()) {
		return func(t *testing.T) (refreshtoken.RefreshTokenRepository, func()) {
			db, cleanup := database_leveldb.Fixture()

			repo, err := leveldb.NewRefreshTokenRepository(db, test.TestPrefix)
			if err != nil {
				t.Fatal(err)
			}

			return repo, cleanup
		}
	})
}
