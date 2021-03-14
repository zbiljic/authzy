package schema

import (
	"time"

	"github.com/zbiljic/authzy/pkg/domain/refreshtoken"
)

type RefreshToken struct {
	ID      string `json:"id" validate:"required,alphanum"`
	UserID  string `json:"user_id" validate:"required,alphanum"`
	Token   string `json:"token" validate:"required,alphanum"`
	Revoked bool   `json:"revoked,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func RefreshTokenToSchema(in *refreshtoken.RefreshToken) *RefreshToken {
	out := &RefreshToken{}
	if in != nil {
		out.ID = in.ID
		out.UserID = in.UserID
		out.Token = in.Token
		out.Revoked = in.Revoked
		out.CreatedAt = in.CreatedAt
		out.UpdatedAt = in.UpdatedAt
	}

	return out
}

func RefreshTokenFromSchema(in *RefreshToken) *refreshtoken.RefreshToken {
	out := &refreshtoken.RefreshToken{}
	out.ID = in.ID
	out.UserID = in.UserID
	out.Token = in.Token
	out.Revoked = in.Revoked
	out.CreatedAt = in.CreatedAt
	out.UpdatedAt = in.UpdatedAt

	return out
}
