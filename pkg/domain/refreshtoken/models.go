package refreshtoken

import "time"

// RefreshToken is the model for refresh tokens.
type RefreshToken struct {
	ID      string
	UserID  string
	Token   string
	Revoked bool

	CreatedAt time.Time
	UpdatedAt time.Time
}
