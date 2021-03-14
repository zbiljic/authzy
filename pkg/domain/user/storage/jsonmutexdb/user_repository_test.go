package jsonmutexdb_test

import (
	"testing"

	"github.com/zbiljic/authzy/pkg/domain/user"
	"github.com/zbiljic/authzy/pkg/domain/user/storage/jsonmutexdb"
	"github.com/zbiljic/authzy/pkg/domain/user/storage/test"
)

func TestJSONMutexDBUserRepository(t *testing.T) {
	test.Run(t, func() func(t *testing.T) (user.UserRepository, func()) {
		return func(t *testing.T) (user.UserRepository, func()) {
			repo, err := jsonmutexdb.NewUserRepository(nil, test.TestPrefix)
			if err != nil {
				t.Fatal(err)
			}

			return repo, func() {}
		}
	})
}
