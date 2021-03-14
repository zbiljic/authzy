package schema

import (
	"github.com/zbiljic/authzy/pkg/domain/account"
)

type Account struct {
	UserID      string `json:"user_id" validate:"required,alphanum"`
	Provider    string `json:"provider" validate:"required"`
	FederatedID string `json:"federated_id" validate:"required"`
}

func (a *Account) BeforeSave() error {
	return nil
}

func AccountToSchema(in *account.Account) *Account {
	out := &Account{}
	if in != nil {
		out.UserID = in.UserID
		out.Provider = in.Provider.String()
		out.FederatedID = in.FederatedID
	}

	return out
}

func AccountFromSchema(in *Account) *account.Account {
	out := &account.Account{}
	out.UserID = in.UserID
	out.Provider = account.ProviderType(in.Provider)
	out.FederatedID = in.FederatedID

	return out
}
