package transformer

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zbiljic/authzy/pkg/domain/account/storage/json/schema"
)

const keySeparator = "/"

const (
	ns                 = "account/storage/json/transformer."
	opMarshalAccount   = ns + "MarshalAccount"
	opUnmarshalAccount = ns + "UnmarshalAccount"
)

func MarshalAccountID(account *schema.Account) string {
	return strings.Join([]string{account.UserID, account.Provider, account.FederatedID}, keySeparator)
}

func MarshalAccountKey(prefix, id string) string {
	return strings.Join([]string{prefix, id}, keySeparator)
}

func MarshalAccount(in *schema.Account) ([]byte, error) {
	out, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", opMarshalAccount, err)
	}

	return out, nil
}

func UnmarshalAccount(in []byte) (*schema.Account, error) {
	out := &schema.Account{}
	err := json.Unmarshal(in, out)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", opUnmarshalAccount, err)
	}

	return out, nil
}
