package leveldb

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
)

// New creates a new local LevelDB database.
func New(config Config) (*leveldb.DB, error) {
	db, err := leveldb.OpenFile(config.DataDir, nil)
	if err != nil {
		return nil, fmt.Errorf("open leveldb: %s", err)
	}

	return db, nil
}
