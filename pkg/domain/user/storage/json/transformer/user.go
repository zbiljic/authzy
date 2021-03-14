package transformer

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zbiljic/authzy/pkg/domain/user/storage/json/schema"
)

const keySeparator = "/"

const (
	ns              = "user/storage/json/transformer."
	opMarshalUser   = ns + "MarshalUser"
	opUnmarshalUser = ns + "UnmarshalUser"
)

func MarshalUserKey(prefix, id string) string {
	return strings.Join([]string{prefix, id}, keySeparator)
}

func MarshalUser(in *schema.User) ([]byte, error) {
	out, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", opMarshalUser, err)
	}

	return out, nil
}

func UnmarshalUser(in []byte) (*schema.User, error) {
	out := &schema.User{}
	err := json.Unmarshal(in, out)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", opUnmarshalUser, err)
	}

	return out, nil
}
