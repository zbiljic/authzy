package di

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
	"go.uber.org/fx"

	database_jsonmutexdb "github.com/zbiljic/authzy/pkg/database/jsonmutexdb"
	database_leveldb "github.com/zbiljic/authzy/pkg/database/leveldb"
	"github.com/zbiljic/authzy/pkg/domain/refreshtoken"
	refreshtoken_jsonmutexdb "github.com/zbiljic/authzy/pkg/domain/refreshtoken/storage/jsonmutexdb"
	refreshtoken_leveldb "github.com/zbiljic/authzy/pkg/domain/refreshtoken/storage/leveldb"
)

var repositoresfx = fx.Provide(
	NewRefreshTokenRepository,
)

type RepositoryParams struct {
	fx.In

	Type string `name:"db_type"`

	JSONMutexDBConfig *database_jsonmutexdb.Config
	LevelDBConfig     *database_leveldb.Config

	JSONLoadSaver *database_jsonmutexdb.LoadSaver `optional:"true"`
	LevelDB       *leveldb.DB                     `optional:"true"`
}

func NewRefreshTokenRepository(p RepositoryParams) (refreshtoken.RefreshTokenRepository, error) {
	switch p.Type {
	case database_jsonmutexdb.Type:
		return NewJSONMutexDBRefreshTokenRepository(p.JSONMutexDBConfig, p.JSONLoadSaver)
	case database_leveldb.Type:
		return NewLevelDBRefreshTokenRepository(p.LevelDBConfig, p.LevelDB)
	default:
		return nil, fmt.Errorf("invalid database type: %s", p.Type)
	}
}

func NewJSONMutexDBRefreshTokenRepository(
	config *database_jsonmutexdb.Config,
	ls *database_jsonmutexdb.LoadSaver,
) (refreshtoken.RefreshTokenRepository, error) {
	return refreshtoken_jsonmutexdb.NewRefreshTokenRepository(
		*ls,
		config.FilenamePrefix,
	)
}

func NewLevelDBRefreshTokenRepository(
	config *database_leveldb.Config,
	db *leveldb.DB,
) (refreshtoken.RefreshTokenRepository, error) {
	return refreshtoken_leveldb.NewRefreshTokenRepository(
		db,
		config.KeyPrefix,
	)
}
