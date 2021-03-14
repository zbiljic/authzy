package di

import (
	"go.uber.org/fx"

	"github.com/zbiljic/authzy/pkg/hash"
)

var hasherfx = fx.Provide(
	ProvideHasher,
)

func ProvideHasher(config *hash.HasherArgon2Config) hash.Hasher {
	return hash.NewHasherArgon2(*config)
}
