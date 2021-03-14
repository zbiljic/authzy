package mock

import (
	"bytes"
	"context"
	"encoding/hex"

	"github.com/zbiljic/authzy/pkg/hash"
)

type mockHasher struct{}

func NewMockHasher() hash.Hasher {
	return &mockHasher{}
}

func (h *mockHasher) Generate(ctx context.Context, password []byte) ([]byte, error) {
	dst := make([]byte, hex.EncodedLen(len(password)))
	hex.Encode(dst, password)

	return dst, nil
}

func (h *mockHasher) Compare(ctx context.Context, password []byte, encrypted []byte) error {
	dst := make([]byte, hex.EncodedLen(len(password)))
	hex.Encode(dst, password)

	if bytes.Equal(encrypted, dst) {
		return nil
	}

	return hash.ErrMismatchedHashAndPassword
}
