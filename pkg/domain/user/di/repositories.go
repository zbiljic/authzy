package di

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
	"go.uber.org/fx"

	database_jsonmutexdb "github.com/zbiljic/authzy/pkg/database/jsonmutexdb"
	database_leveldb "github.com/zbiljic/authzy/pkg/database/leveldb"
	"github.com/zbiljic/authzy/pkg/domain/user"
	user_jsonmutexdb "github.com/zbiljic/authzy/pkg/domain/user/storage/jsonmutexdb"
	user_leveldb "github.com/zbiljic/authzy/pkg/domain/user/storage/leveldb"
)

var repositoresfx = fx.Provide(
	NewUserRepository,
)

type RepositoryParams struct {
	fx.In

	Type string `name:"db_type"`

	JSONMutexDBConfig *database_jsonmutexdb.Config
	LevelDBConfig     *database_leveldb.Config

	JSONLoadSaver *database_jsonmutexdb.LoadSaver `optional:"true"`
	LevelDB       *leveldb.DB                     `optional:"true"`
}

func NewUserRepository(p RepositoryParams) (user.UserRepository, error) {
	switch p.Type {
	case database_jsonmutexdb.Type:
		return NewJSONMutexDBUserRepository(p.JSONMutexDBConfig, p.JSONLoadSaver)
	case database_leveldb.Type:
		return NewLevelDBUserRepository(p.LevelDBConfig, p.LevelDB)
	default:
		return nil, fmt.Errorf("invalid database type: %s", p.Type)
	}
}

func NewJSONMutexDBUserRepository(
	config *database_jsonmutexdb.Config,
	ls *database_jsonmutexdb.LoadSaver,
) (user.UserRepository, error) {
	return user_jsonmutexdb.NewUserRepository(
		*ls,
		config.FilenamePrefix,
	)
}

func NewLevelDBUserRepository(
	config *database_leveldb.Config,
	db *leveldb.DB,
) (user.UserRepository, error) {
	return user_leveldb.NewUserRepository(
		db,
		config.KeyPrefix,
	)
}
