package transformer

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zbiljic/authzy/pkg/domain/refreshtoken/storage/json/schema"
)

const keySeparator = "/"

const (
	ns                      = "refreshtoken/storage/json/transformer."
	opMarshalRefreshToken   = ns + "MarshalRefreshToken"
	opUnmarshalRefreshToken = ns + "UnmarshalRefreshToken"
)

func MarshalRefreshTokenKey(prefix, id string) string {
	return strings.Join([]string{prefix, id}, keySeparator)
}

func MarshalRefreshTokenUserIDKey(prefix, userID, id string) string {
	return strings.Join([]string{prefix, userID, id}, keySeparator)
}

func UnmarshalRefreshTokenUserIDKey(key string) (userID, id string) {
	split := strings.Split(key, keySeparator)
	return split[len(split)-2], split[len(split)-1]
}

func MarshalRefreshToken(in *schema.RefreshToken) ([]byte, error) {
	out, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", opMarshalRefreshToken, err)
	}

	return out, nil
}

func UnmarshalRefreshToken(in []byte) (*schema.RefreshToken, error) {
	out := &schema.RefreshToken{}
	err := json.Unmarshal(in, out)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", opUnmarshalRefreshToken, err)
	}

	return out, nil
}
