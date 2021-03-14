package di

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
	"go.uber.org/fx"

	database_jsonmutexdb "github.com/zbiljic/authzy/pkg/database/jsonmutexdb"
	database_leveldb "github.com/zbiljic/authzy/pkg/database/leveldb"
)

var databasefx = fx.Provide(
	ProvideDatabase,
)

type DatabaseParams struct {
	fx.In

	Type string `name:"db_type"`

	JSONMutexDBConfig *database_jsonmutexdb.Config
	LevelDBConfig     *database_leveldb.Config
}

type DatabaseResult struct {
	fx.Out

	JSONLoadSaver *database_jsonmutexdb.LoadSaver `optional:"true"`
	LevelDB       *leveldb.DB                     `optional:"true"`
}

func ProvideDatabase(p DatabaseParams) (DatabaseResult, error) {
	result := DatabaseResult{}

	switch p.Type {
	case database_jsonmutexdb.Type:
		ls, err := ProvideJSONMutexDBLoadSaver(p.JSONMutexDBConfig)
		if err != nil {
			return result, err
		}
		result.JSONLoadSaver = ls
		return result, nil
	case database_leveldb.Type:
		ldb, err := ProvideLevelDBDatabase(p.LevelDBConfig)
		if err != nil {
			return result, err
		}
		result.LevelDB = ldb
		return result, nil
	default:
		return result, fmt.Errorf("invalid database type: %s", p.Type)
	}
}

func ProvideJSONMutexDBLoadSaver(config *database_jsonmutexdb.Config) (*database_jsonmutexdb.LoadSaver, error) {
	ls, err := database_jsonmutexdb.NewLoadSaver(config.DataDir)
	if err != nil {
		return nil, err
	}

	return &ls, nil
}

func ProvideLevelDBDatabase(config *database_leveldb.Config) (*leveldb.DB, error) {
	return database_leveldb.New(*config)
}
