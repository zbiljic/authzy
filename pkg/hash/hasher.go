package hash

import "context"

// Hasher provides methods for generating and comparing password hashes.
type Hasher interface {
	// Compare a password to a hash and return nil if they match or an error otherwise.
	Compare(ctx context.Context, password []byte, hash []byte) error

	// Generate returns a hash derived from the password or an error if the hash method failed.
	Generate(ctx context.Context, password []byte) ([]byte, error)
}

type HashProvider interface {
	Hasher() Hasher
}
