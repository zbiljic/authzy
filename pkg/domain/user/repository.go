package user

import "context"

type UserRepository interface {
	// Save saves a given entity.
	Save(ctx context.Context, entity *User) (*User, error)

	// FindByID retrieves an entity by its id.
	FindByID(ctx context.Context, id string) (*User, error)

	// ExistsByID returns whether an entity with the given id exists.
	ExistsByID(ctx context.Context, id string) (bool, error)

	// FindAll returns all instances of the type.
	FindAll(ctx context.Context, afterCursor string, limit int) ([]*User, string, error)

	// Count returns the number of entities available.
	Count(ctx context.Context) (int, error)

	// DeleteByID deletes the entity with the given id.
	DeleteByID(ctx context.Context, id string) error

	// Delete deletes a given entity.
	Delete(ctx context.Context, entity *User) error

	// DeleteAll deletes all entities managed by the repository.
	DeleteAll(ctx context.Context) error

	// ExistsByIdentifier returns whether an user with the given identifier
	// (username OR email) exists.
	ExistsByIdentifier(ctx context.Context, identifier string) (bool, error)

	// FindByIdentifier retrieves an user by its identifier (username OR email).
	FindByIdentifier(ctx context.Context, identifier string) (*User, error)

	// FindByConfirmationToken finds user with the matching confirmation token.
	FindByConfirmationToken(ctx context.Context, token string) (*User, error)

	// FindByRecoveryToken finds user with the matching recovery token.
	FindByRecoveryToken(ctx context.Context, token string) (*User, error)
}
