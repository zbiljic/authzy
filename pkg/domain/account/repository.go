package account

import "context"

type AccountRepository interface {
	// Save saves a given entity.
	Save(ctx context.Context, entity *Account) (*Account, error)

	// Find retrieves an entity.
	Find(ctx context.Context, entity *Account) (*Account, error)

	// Exists returns whether an entity exists.
	Exists(ctx context.Context, entity *Account) (bool, error)

	// FindAll returns all instances of the type.
	FindAll(ctx context.Context, afterCursor string, limit int) ([]*Account, string, error)

	// Count returns the number of entities available.
	Count(ctx context.Context) (int, error)

	// Delete deletes a given entity.
	Delete(ctx context.Context, entity *Account) error

	// DeleteAll deletes all entities managed by the repository.
	DeleteAll(ctx context.Context) error

	// FindAllForUser returns all refresh tokens for specified user ID.
	FindAllForUser(ctx context.Context, userID string) ([]*Account, error)
}
