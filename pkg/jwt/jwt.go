package jwt

import (
	"fmt"
	"time"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"

	"github.com/zbiljic/authzy/pkg/config"
)

type Service interface {
	Generate(string) (jwt.Token, error)
	Sign(jwt.Token) (string, error)
	Parse([]byte) (jwt.Token, error)
	Validate(jwt.Token) error
}

type jwtService struct {
	audience       string
	expireAfter    time.Duration
	acceptableSkew time.Duration
	defaultKey     jwk.Key
	keyset         jwk.Set
}

func NewService(jwtConfig *config.JWTConfig) (Service, error) {
	keyset, err := jwk.Parse([]byte(jwtConfig.KeysJSON))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT key set: %w", err)
	}

	defaultKey, ok := keyset.LookupKeyID(jwtConfig.DefaultKey)
	if !ok {
		return nil, fmt.Errorf("default JWT key missing: %v", jwtConfig.DefaultKey)
	}

	return &jwtService{
		audience:       jwtConfig.Aud,
		expireAfter:    time.Second * time.Duration(jwtConfig.Exp),
		acceptableSkew: jwtConfig.AcceptableSkew,
		defaultKey:     defaultKey,
		keyset:         keyset,
	}, nil
}

//nolint:errcheck
func (s *jwtService) Generate(subject string) (jwt.Token, error) {
	now := time.Now()

	t := jwt.New()

	t.Set(jwt.SubjectKey, subject)
	if s.audience != "" {
		t.Set(jwt.AudienceKey, s.audience)
	}
	t.Set(jwt.IssuedAtKey, now)
	t.Set(jwt.ExpirationKey, now.Add(s.expireAfter))
	t.Set(jwt.NotBeforeKey, now)

	return t, nil
}

func (s *jwtService) Sign(token jwt.Token) (string, error) {
	key := s.defaultKey

	token.Set(jwt.JwtIDKey, key.KeyID()) //nolint:errcheck

	signed, err := jwt.Sign(token, jwa.HS256, key)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %s", err)
	}

	return string(signed), nil
}

func (s *jwtService) Parse(payload []byte) (jwt.Token, error) {
	token, err := jwt.Parse(payload,
		jwt.WithKeySet(s.keyset),
		jwt.WithValidate(false),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse payload: %w", err)
	}

	return token, nil
}

func (s *jwtService) Validate(token jwt.Token) error {
	err := jwt.Validate(token,
		jwt.WithAcceptableSkew(s.acceptableSkew),
	)
	if err != nil {
		return fmt.Errorf("failed to validate token: %w", err)
	}

	return nil
}
