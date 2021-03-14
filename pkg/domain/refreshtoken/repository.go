package refreshtoken

import "context"

type RefreshTokenRepository interface {
	// Save saves a given entity.
	Save(ctx context.Context, entity *RefreshToken) (*RefreshToken, error)

	// FindByToken retrieves an entity by its id.
	FindByID(ctx context.Context, id string) (*RefreshToken, error)

	// ExistsByID returns whether an entity with the given id exists.
	ExistsByID(ctx context.Context, id string) (bool, error)

	// FindAll returns all instances of the type.
	FindAll(ctx context.Context, afterCursor string, limit int) ([]*RefreshToken, string, error)

	// Count Returns the number of entities available.
	Count(ctx context.Context) (int, error)

	// DeleteByID deletes the entity with the given id.
	DeleteByID(ctx context.Context, id string) error

	// Delete deletes a given entity.
	Delete(ctx context.Context, entity *RefreshToken) error

	// DeleteAll deletes all entities managed by the repository.
	DeleteAll(ctx context.Context) error

	// FindByToken retrieves an entity by token value.
	FindByToken(ctx context.Context, token string) (*RefreshToken, error)

	// FindAllForUser returns all refresh tokens for specified user ID.
	FindAllForUser(ctx context.Context, userID, afterCursor string, limit int) ([]*RefreshToken, string, error)
}
