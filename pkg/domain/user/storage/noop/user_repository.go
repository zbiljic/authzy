package noop

import (
	"context"

	"github.com/zbiljic/authzy/pkg/domain/user"
)

// Compile-time proof of interface implementation.
var _ user.UserRepository = (*UnimplementedUserRepository)(nil)

// UnimplementedUserRepository can be embedded to have forward compatible implementations.
type UnimplementedUserRepository struct{}

func (*UnimplementedUserRepository) Save(ctx context.Context, entity *user.User) (*user.User, error) {
	panic("Save not implemented")
}

func (*UnimplementedUserRepository) FindByID(ctx context.Context, id string) (*user.User, error) {
	panic("FindByID not implemented")
}

func (*UnimplementedUserRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	panic("ExistsByID not implemented")
}

func (*UnimplementedUserRepository) FindAll(ctx context.Context, afterCursor string, limit int) ([]*user.User, string, error) {
	panic("FindAll not implemented")
}

func (*UnimplementedUserRepository) Count(ctx context.Context) (int, error) {
	panic("Count not implemented")
}

func (*UnimplementedUserRepository) DeleteByID(ctx context.Context, id string) error {
	panic("DeleteByID not implemented")
}

func (*UnimplementedUserRepository) Delete(ctx context.Context, entity *user.User) error {
	panic("Delete not implemented")
}

func (*UnimplementedUserRepository) DeleteAll(ctx context.Context) error {
	panic("DeleteAll not implemented")
}

func (*UnimplementedUserRepository) ExistsByIdentifier(ctx context.Context, identifier string) (bool, error) {
	panic("ExistsByIdentifier not implemented")
}

func (*UnimplementedUserRepository) FindByIdentifier(ctx context.Context, identifier string) (*user.User, error) {
	panic("FindByIdentifier not implemented")
}

func (*UnimplementedUserRepository) FindByConfirmationToken(ctx context.Context, token string) (*user.User, error) {
	panic("FindByConfirmationToken not implemented")
}

func (*UnimplementedUserRepository) FindByRecoveryToken(ctx context.Context, token string) (*user.User, error) {
	panic("FindByRecoveryToken not implemented")
}
