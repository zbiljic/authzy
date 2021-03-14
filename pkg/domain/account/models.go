package account

import (
	"encoding/json"
	"errors"
)

// Account respresents single account for the user.
type Account struct {
	UserID      string
	Provider    ProviderType
	FederatedID string
}

type ProviderType string

const (
	ProviderTypePassword ProviderType = "password"
)

func (p *ProviderType) UnmarshalJSON(b []byte) error {
	var s string
	json.Unmarshal(b, &s) //nolint:errcheck
	leaveType := ProviderType(s)
	switch leaveType {
	case ProviderTypePassword:
		*p = leaveType
		return nil
	}
	return errors.New("Invalid provider type")
}

func (p ProviderType) IsValid() error {
	switch p {
	case ProviderTypePassword:
		return nil
	}
	return errors.New("Invalid provider type")
}

func (p ProviderType) String() string {
	return string(p)
}
