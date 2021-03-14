package di

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
	"go.uber.org/fx"

	database_jsonmutexdb "github.com/zbiljic/authzy/pkg/database/jsonmutexdb"
	database_leveldb "github.com/zbiljic/authzy/pkg/database/leveldb"
	"github.com/zbiljic/authzy/pkg/domain/account"
	account_jsonmutexdb "github.com/zbiljic/authzy/pkg/domain/account/storage/jsonmutexdb"
	account_leveldb "github.com/zbiljic/authzy/pkg/domain/account/storage/leveldb"
)

var repositoresfx = fx.Provide(
	NewAccountRepository,
)

type RepositoryParams struct {
	fx.In

	Type string `name:"db_type"`

	JSONMutexDBConfig *database_jsonmutexdb.Config
	LevelDBConfig     *database_leveldb.Config

	JSONLoadSaver *database_jsonmutexdb.LoadSaver `optional:"true"`
	LevelDB       *leveldb.DB                     `optional:"true"`
}

func NewAccountRepository(p RepositoryParams) (account.AccountRepository, error) {
	switch p.Type {
	case database_jsonmutexdb.Type:
		return NewJSONMutexDBAccountRepository(p.JSONMutexDBConfig, p.JSONLoadSaver)
	case database_leveldb.Type:
		return NewLevelDBAccountRepository(p.LevelDBConfig, p.LevelDB)
	default:
		return nil, fmt.Errorf("invalid database type: %s", p.Type)
	}
}

func NewJSONMutexDBAccountRepository(
	config *database_jsonmutexdb.Config,
	ls *database_jsonmutexdb.LoadSaver,
) (account.AccountRepository, error) {
	return account_jsonmutexdb.NewAccountRepository(
		*ls,
		config.FilenamePrefix,
	)
}

func NewLevelDBAccountRepository(
	config *database_leveldb.Config,
	db *leveldb.DB,
) (account.AccountRepository, error) {
	return account_leveldb.NewAccountRepository(
		db,
		config.KeyPrefix,
	)
}
