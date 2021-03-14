package noop

import (
	"context"

	"github.com/zbiljic/authzy/pkg/domain/account"
)

// Compile-time proof of interface implementation.
var _ account.AccountRepository = (*UnimplementedAccountRepository)(nil)

// UnimplementedAccountRepository can be embedded to have forward compatible implementations.
type UnimplementedAccountRepository struct{}

func (*UnimplementedAccountRepository) Save(ctx context.Context, entity *account.Account) (*account.Account, error) {
	panic("Save not implemented")
}

func (*UnimplementedAccountRepository) Find(ctx context.Context, entity *account.Account) (*account.Account, error) {
	panic("Find not implemented")
}

func (*UnimplementedAccountRepository) Exists(ctx context.Context, entity *account.Account) (bool, error) {
	panic("Exists not implemented")
}

func (*UnimplementedAccountRepository) FindAll(ctx context.Context, afterCursor string, limit int) ([]*account.Account, string, error) {
	panic("FindAll not implemented")
}

func (*UnimplementedAccountRepository) Count(ctx context.Context) (int, error) {
	panic("Count not implemented")
}

func (*UnimplementedAccountRepository) Delete(ctx context.Context, entity *account.Account) error {
	panic("Delete not implemented")
}

func (*UnimplementedAccountRepository) DeleteAll(ctx context.Context) error {
	panic("DeleteAll not implemented")
}

func (*UnimplementedAccountRepository) FindAllForUser(ctx context.Context, userID string) ([]*account.Account, error) {
	panic("FindAllForUser not implemented")
}
