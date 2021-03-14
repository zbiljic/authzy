package leveldb

import (
	"os"

	"github.com/syndtr/goleveldb/leveldb"

	"github.com/zbiljic/authzy/pkg/testutil"
)

// Fixture returns a temporary test database for testing.
func Fixture() (*leveldb.DB, func()) {
	var cleanup testutil.Cleanup
	defer cleanup.Recover()

	tmpdir, err := os.MkdirTemp(".", "test-db-")
	if err != nil {
		panic(err)
	}
	cleanup.Add(func() { os.RemoveAll(tmpdir) })

	db, err := New(Config{DataDir: tmpdir})
	if err != nil {
		panic(err)
	}

	return db, cleanup.Run
}
