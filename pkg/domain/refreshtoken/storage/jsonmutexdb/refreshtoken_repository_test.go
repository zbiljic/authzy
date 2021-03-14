package jsonmutexdb_test

import (
	"testing"

	"github.com/zbiljic/authzy/pkg/domain/refreshtoken"
	"github.com/zbiljic/authzy/pkg/domain/refreshtoken/storage/jsonmutexdb"
	"github.com/zbiljic/authzy/pkg/domain/refreshtoken/storage/test"
)

func TestJSONMutexDBRefreshTokenRepository(t *testing.T) {
	test.Run(t, func() func(t *testing.T) (refreshtoken.RefreshTokenRepository, func()) {
		return func(t *testing.T) (refreshtoken.RefreshTokenRepository, func()) {
			repo, err := jsonmutexdb.NewRefreshTokenRepository(nil, test.TestPrefix)
			if err != nil {
				t.Fatal(err)
			}

			return repo, func() {}
		}
	})
}
