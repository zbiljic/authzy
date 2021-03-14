package di

import (
	"go.uber.org/fx"

	"github.com/zbiljic/authzy/pkg/config"
	database_jsonmutexdb "github.com/zbiljic/authzy/pkg/database/jsonmutexdb"
	database_leveldb "github.com/zbiljic/authzy/pkg/database/leveldb"
	"github.com/zbiljic/authzy/pkg/hash"
	"github.com/zbiljic/authzy/pkg/logger"
)

var configfx = fx.Provide(
	ProvideLoggerConfig,
	ProvideDebugConfig,
	ProvideHTTPConfig,
	ProvideHashersConfig,
	ProvideHasherArgon2Config,
	ProvideDatabaseConfig,
	ProvideDatabaseConfigResult,
	ProvideAPIConfig,
	ProvideAPIJWTConfig,
)

func ProvideLoggerConfig(config *config.Config) *logger.Config {
	return config.Logger
}

func ProvideDebugConfig(config *config.Config) *config.DebugConfig {
	return config.Debug
}

func ProvideHTTPConfig(config *config.Config) *config.HTTPConfig {
	return config.HTTP
}

func ProvideHashersConfig(config *config.Config) *config.HashersConfig {
	return config.Hashers
}

func ProvideHasherArgon2Config(config *config.HashersConfig) *hash.HasherArgon2Config {
	return config.Argon2
}

func ProvideDatabaseConfig(config *config.Config) *config.DatabaseConfig {
	return config.Database
}

type DatabaseConfigResult struct {
	fx.Out

	Type              string `name:"db_type"`
	JSONMutexDBConfig *database_jsonmutexdb.Config
	LevelDBConfig     *database_leveldb.Config
}

func ProvideDatabaseConfigResult(config *config.DatabaseConfig) DatabaseConfigResult {
	return DatabaseConfigResult{
		Type:              config.Type,
		JSONMutexDBConfig: config.JSONMutexDB,
		LevelDBConfig:     config.LevelDB,
	}
}

func ProvideAPIConfig(config *config.Config) *config.APIConfig {
	return config.API
}

func ProvideAPIJWTConfig(config *config.APIConfig) *config.JWTConfig {
	return config.JWT
}
