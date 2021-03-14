package noop

import (
	"context"

	"github.com/zbiljic/authzy/pkg/domain/refreshtoken"
)

// Compile-time proof of interface implementation.
var _ refreshtoken.RefreshTokenRepository = (*UnimplementedRefreshTokenRepository)(nil)

// UnimplementedRefreshTokenRepository can be embedded to have forward compatible implementations.
type UnimplementedRefreshTokenRepository struct{}

func (*UnimplementedRefreshTokenRepository) Save(ctx context.Context, entity *refreshtoken.RefreshToken) (*refreshtoken.RefreshToken, error) {
	panic("Save not implemented")
}

func (*UnimplementedRefreshTokenRepository) FindByID(ctx context.Context, id string) (*refreshtoken.RefreshToken, error) {
	panic("FindByID not implemented")
}

func (*UnimplementedRefreshTokenRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	panic("ExistsByID not implemented")
}

func (*UnimplementedRefreshTokenRepository) FindAll(ctx context.Context, afterCursor string, limit int) ([]*refreshtoken.RefreshToken, string, error) {
	panic("FindAll not implemented")
}

func (*UnimplementedRefreshTokenRepository) Count(ctx context.Context) (int, error) {
	panic("Count not implemented")
}

func (*UnimplementedRefreshTokenRepository) DeleteByID(ctx context.Context, id string) error {
	panic("DeleteByID not implemented")
}

func (*UnimplementedRefreshTokenRepository) Delete(ctx context.Context, entity *refreshtoken.RefreshToken) error {
	panic("Delete not implemented")
}

func (*UnimplementedRefreshTokenRepository) DeleteAll(ctx context.Context) error {
	panic("DeleteAll not implemented")
}

func (*UnimplementedRefreshTokenRepository) FindByToken(ctx context.Context, token string) (*refreshtoken.RefreshToken, error) {
	panic("FindByToken not implemented")
}

func (*UnimplementedRefreshTokenRepository) FindAllForUser(ctx context.Context, userID, afterCursor string, limit int) ([]*refreshtoken.RefreshToken, string, error) {
	panic("FindAllForUser not implemented")
}
