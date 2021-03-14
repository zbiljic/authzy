package hash

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidHash               = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion       = errors.New("incompatible version of argon2")
	ErrMismatchedHashAndPassword = errors.New("passwords do not match")
)

const (
	Argon2DefaultMemory      uint32 = 4 * 1024 * 1024
	Argon2DefaultIterations  uint32 = 4
	Argon2DefaultParallelism uint8  = 8
	Argon2DefaultSaltLength  uint32 = 16
	Argon2DefaultKeyLength   uint32 = 32
)

var Argon2CPUParallelism = uint8(runtime.NumCPU() * 2)

type HasherArgon2Config struct {
	Memory      uint32 `json:"memory" required:"true" default:"4194304"`
	Iterations  uint32 `json:"iterations" required:"true" default:"4"`
	Parallelism uint8  `json:"parallelism" required:"true" default:"8"`
	SaltLength  uint32 `json:"salt_length" required:"true" default:"16" split_words:"true"`
	KeyLength   uint32 `json:"key_length" required:"true" default:"32" split_words:"true"`
}

type Argon2 struct {
	c HasherArgon2Config
}

func NewHasherArgon2(c HasherArgon2Config) *Argon2 {
	return &Argon2{c: c}
}

func (h *Argon2) Generate(ctx context.Context, password []byte) ([]byte, error) {
	p := h.c

	salt := make([]byte, p.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	// Pass the plaintext password, salt and parameters to the argon2.IDKey
	// function. This will generate a hash of the password using the Argon2id
	// variant.
	hash := argon2.IDKey(password, salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)

	var b bytes.Buffer
	if _, err := fmt.Fprintf(
		&b,
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, p.Memory, p.Iterations, p.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	); err != nil {
		return nil, errors.WithStack(err)
	}

	return b.Bytes(), nil
}

func (h *Argon2) Compare(ctx context.Context, password []byte, hash []byte) error {
	// Extract the parameters, salt and derived key from the encoded password
	// hash.
	p, salt, hash, err := decodeHash(string(hash))
	if err != nil {
		return err
	}

	// Derive the key from the other password using the same parameters.
	otherHash := argon2.IDKey(password, salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)

	// Check that the contents of the hashed passwords are identical. Note
	// that we are using the subtle.ConstantTimeCompare() function for this
	// to help prevent timing attacks.
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return nil
	}

	return ErrMismatchedHashAndPassword
}

func decodeHash(encodedHash string) (p *HasherArgon2Config, salt, hash []byte, err error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, ErrInvalidHash
	}

	var version int
	_, err = fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	p = new(HasherArgon2Config)
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Iterations, &p.Parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, err
	}
	p.SaltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, err
	}
	p.KeyLength = uint32(len(hash))

	return p, salt, hash, nil
}
