package leveldb

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/zbiljic/authzy/pkg/database"
	"github.com/zbiljic/authzy/pkg/domain/user"
	"github.com/zbiljic/authzy/pkg/domain/user/storage/json/schema"
	"github.com/zbiljic/authzy/pkg/domain/user/storage/json/transformer"
	"github.com/zbiljic/authzy/pkg/domain/user/storage/noop"
)

const (
	usersPrefix                       = "users"
	usersIdentifierIndexPrefix        = "index_users_identifier"
	usersConfirmationTokenIndexPrefix = "index_users_confirmation_token"
	usersRecoveryTokenIndexPrefix     = "index_users_recovery_token"
)

// levelDBUserRepository is a repository that uses LevelDB database.
type levelDBUserRepository struct {
	noop.UnimplementedUserRepository

	db *leveldb.DB
	mu sync.Mutex

	usersKeyspace                       string
	usersIdentifierIndexKeyspace        string
	usersConfirmationTokenIndexKeyspace string
	usersRecoveryTokenIndexKeyspace     string

	validate *validator.Validate
}

// NewUserRepository returns a new LevelDB repository.
func NewUserRepository(
	db *leveldb.DB,
	keyPrefix string,
) (user.UserRepository, error) {
	validate := validator.New()
	schema.RegisterValidators(validate)

	r := &levelDBUserRepository{
		db:                                  db,
		usersKeyspace:                       keyPrefix + usersPrefix,
		usersIdentifierIndexKeyspace:        keyPrefix + usersIdentifierIndexPrefix,
		usersConfirmationTokenIndexKeyspace: keyPrefix + usersConfirmationTokenIndexPrefix,
		usersRecoveryTokenIndexKeyspace:     keyPrefix + usersRecoveryTokenIndexPrefix,
		validate:                            validate,
	}

	return r, nil
}

const (
	ns                        = "user/storage/leveldb."
	opSave                    = ns + "Save"
	opFindByID                = ns + "FindByID"
	opExistsByID              = ns + "ExistsByID"
	opFindAll                 = ns + "FindAll"
	opCount                   = ns + "Count"
	opDeleteByID              = ns + "DeleteByID"
	opExistsByIdentifier      = ns + "ExistsByIdentifier"
	opFindByIdentifier        = ns + "FindByIdentifier"
	opFindByConfirmationToken = ns + "FindByConfirmationToken"
	opFindByRecoveryToken     = ns + "FindByRecoveryToken"
)

func (r *levelDBUserRepository) commit(ctx context.Context, batch *leveldb.Batch) error {
	return r.db.Write(batch, nil)
}

func (r *levelDBUserRepository) Save(ctx context.Context, entity *user.User) (*user.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	inS := schema.UserToSchema(entity)

	err := r.validate.Struct(inS)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", opSave, err)
	}

	// before save
	err = inS.BeforeSave()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", opSave, err)
	}

	if inS.CreatedAt.IsZero() {
		inS.CreatedAt = time.Now()
	}
	inS.UpdatedAt = time.Now()

	key := transformer.MarshalUserKey(r.usersKeyspace, inS.ID)

	value, err := transformer.MarshalUser(inS)
	if err != nil {
		return nil, fmt.Errorf("%s(%s): %w", opSave, inS.ID, err)
	}

	batch := new(leveldb.Batch)

	batch.Put([]byte(key), value)

	if inS.PasswordHash != "" {
		// put base fields only in index
		usernameKey := transformer.MarshalUserKey(r.usersIdentifierIndexKeyspace, inS.NormalizedUsername)
		emailKey := transformer.MarshalUserKey(r.usersIdentifierIndexKeyspace, inS.Email)

		partialUser := schema.User{
			ID:           inS.ID,
			PasswordHash: inS.PasswordHash,
		}

		partialValue, err := transformer.MarshalUser(&partialUser)
		if err != nil {
			return nil, fmt.Errorf("%s(%s): %w", opSave, inS.ID, err)
		}

		batch.Put([]byte(usernameKey), partialValue)
		batch.Put([]byte(emailKey), partialValue)
	}

	ctKey := transformer.MarshalUserKey(r.usersConfirmationTokenIndexKeyspace, inS.ConfirmationToken)
	if inS.ConfirmationToken != "" {
		partialUser := schema.User{
			ID:                 inS.ID,
			ConfirmationToken:  inS.ConfirmationToken,
			ConfirmationSentAt: inS.ConfirmationSentAt,
		}

		partialValue, err := transformer.MarshalUser(&partialUser)
		if err != nil {
			return nil, fmt.Errorf("%s(%s): %w", opSave, inS.ID, err)
		}

		batch.Put([]byte(ctKey), partialValue)
	} else {
		has, err := r.db.Has([]byte(ctKey), nil)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", opSave, err)
		}

		if has {
			batch.Delete([]byte(ctKey))
		}
	}

	rtKey := transformer.MarshalUserKey(r.usersRecoveryTokenIndexKeyspace, inS.RecoveryToken)
	if inS.RecoveryToken != "" {
		partialUser := schema.User{
			ID:             inS.ID,
			RecoveryToken:  inS.RecoveryToken,
			RecoverySentAt: inS.RecoverySentAt,
		}

		partialValue, err := transformer.MarshalUser(&partialUser)
		if err != nil {
			return nil, fmt.Errorf("%s(%s): %w", opSave, inS.ID, err)
		}

		batch.Put([]byte(rtKey), partialValue)
	} else {
		has, err := r.db.Has([]byte(rtKey), nil)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", opSave, err)
		}

		if has {
			batch.Delete([]byte(rtKey))
		}
	}

	err = r.commit(ctx, batch)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", opSave, err)
	}

	savedEntity := schema.UserFromSchema(inS)

	return savedEntity, nil
}

func (r *levelDBUserRepository) FindByID(ctx context.Context, id string) (*user.User, error) {
	key := transformer.MarshalUserKey(r.usersKeyspace, id)

	value, err := r.db.Get([]byte(key), nil)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return nil, fmt.Errorf("%s(%s): %w", opFindByID, id, database.ErrNotFound)
		}

		return nil, fmt.Errorf("%s(%s): %w", opFindByID, id, err)
	}

	ts, err := transformer.UnmarshalUser(value)
	if err != nil {
		return nil, fmt.Errorf("%s(%s): %w", opFindByID, id, err)
	}

	entity := schema.UserFromSchema(ts)

	return entity, nil
}

func (r *levelDBUserRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	key := transformer.MarshalUserKey(r.usersKeyspace, id)

	has, err := r.db.Has([]byte(key), nil)
	if err != nil {
		return false, fmt.Errorf("%s(%s): %w", opExistsByID, id, err)
	}

	return has, nil
}

func (r *levelDBUserRepository) FindAll(ctx context.Context, afterCursor string, limit int) ([]*user.User, string, error) {
	if limit <= 0 {
		limit = 25
	}

	var (
		offset     int
		result     []*user.User
		nextCursor string
	)

	iter := r.db.NewIterator(util.BytesPrefix([]byte(r.usersKeyspace)), nil)
	defer iter.Release()

	if afterCursor != "" {
		key := transformer.MarshalUserKey(r.usersKeyspace, afterCursor)

		if ok := iter.Seek([]byte(key)); !ok {
			err := iter.Error()
			if err != nil {
				return nil, "", fmt.Errorf("%s(%s): %w", opFindAll, afterCursor, err)
			}
		}
	}

	for iter.Next() {
		offset++

		ts, err := transformer.UnmarshalUser(iter.Value())
		if err != nil {
			return nil, "", fmt.Errorf("%s(%s): %w", opFindAll, string(iter.Key()), err)
		}

		t := schema.UserFromSchema(ts)

		result = append(result, t)

		if limit == offset {
			break // stops iterator
		}
	}

	err := iter.Error()
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", opFindAll, err)
	}

	if len(result) == limit {
		nextCursor = result[len(result)-1].ID
	}

	return result, nextCursor, nil
}

func (r *levelDBUserRepository) Count(ctx context.Context) (int, error) {
	var (
		count          int
		ctxCheckOffset int
	)

	iter := r.db.NewIterator(util.BytesPrefix([]byte(r.usersKeyspace)), nil)
	defer iter.Release()

	for iter.Next() {
		if count == ctxCheckOffset {
			select {
			case <-ctx.Done():
				return count, fmt.Errorf("%s: %w", opCount, ctx.Err())
			default:
			}

			ctxCheckOffset += 100
		}

		count++
	}

	err := iter.Error()
	if err != nil {
		return count, fmt.Errorf("%s: %w", opCount, err)
	}

	return count, nil
}

func (r *levelDBUserRepository) DeleteByID(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := transformer.MarshalUserKey(r.usersKeyspace, id)

	value, err := r.db.Get([]byte(key), nil)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return nil
		}

		return fmt.Errorf("%s(%s): %w", opDeleteByID, id, err)
	}

	ts, err := transformer.UnmarshalUser(value)
	if err != nil {
		return fmt.Errorf("%s(%s): %w", opDeleteByID, id, err)
	}

	entity := schema.UserFromSchema(ts)

	batch := new(leveldb.Batch)

	// delete main value
	batch.Delete([]byte(key))

	// delete from index
	usernameKey := transformer.MarshalUserKey(r.usersIdentifierIndexKeyspace, entity.NormalizedUsername)
	emailKey := transformer.MarshalUserKey(r.usersIdentifierIndexKeyspace, entity.Email)

	batch.Delete([]byte(usernameKey))
	batch.Delete([]byte(emailKey))

	if entity.ConfirmationToken != "" {
		ctKey := transformer.MarshalUserKey(r.usersConfirmationTokenIndexKeyspace, entity.ConfirmationToken)
		batch.Delete([]byte(ctKey))
	}

	if entity.RecoveryToken != "" {
		rtKey := transformer.MarshalUserKey(r.usersRecoveryTokenIndexKeyspace, entity.RecoveryToken)
		batch.Delete([]byte(rtKey))
	}

	err = r.commit(ctx, batch)
	if err != nil {
		return fmt.Errorf("%s: %w", opDeleteByID, err)
	}

	return nil
}

func (r *levelDBUserRepository) ExistsByIdentifier(ctx context.Context, identifier string) (bool, error) {
	identifierKey := transformer.MarshalUserKey(r.usersIdentifierIndexKeyspace, identifier)

	has, err := r.db.Has([]byte(identifierKey), nil)
	if err != nil {
		return false, fmt.Errorf("%s(%s): %w", opExistsByIdentifier, identifier, err)
	}

	return has, nil
}

func (r *levelDBUserRepository) FindByIdentifier(ctx context.Context, identifier string) (*user.User, error) {
	// find using index first
	identifierKey := transformer.MarshalUserKey(r.usersIdentifierIndexKeyspace, identifier)

	identifierValue, err := r.db.Get([]byte(identifierKey), nil)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return nil, fmt.Errorf("%s(%s): %w", opFindByIdentifier, identifier, database.ErrNotFound)
		}

		return nil, fmt.Errorf("%s(%s): %w", opFindByIdentifier, identifier, err)
	}

	identifierTs, err := transformer.UnmarshalUser(identifierValue)
	if err != nil {
		return nil, fmt.Errorf("%s(%s): %w", opFindByIdentifier, identifier, err)
	}

	return r.FindByID(ctx, identifierTs.ID)
}

func (r *levelDBUserRepository) FindByConfirmationToken(ctx context.Context, token string) (*user.User, error) {
	ctKey := transformer.MarshalUserKey(r.usersConfirmationTokenIndexKeyspace, token)

	ctValue, err := r.db.Get([]byte(ctKey), nil)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return nil, fmt.Errorf("%s(%s): %w", opFindByConfirmationToken, token, database.ErrNotFound)
		}

		return nil, fmt.Errorf("%s(%s): %w", opFindByConfirmationToken, token, err)
	}

	ctTs, err := transformer.UnmarshalUser(ctValue)
	if err != nil {
		return nil, fmt.Errorf("%s(%s): %w", opFindByConfirmationToken, token, err)
	}

	return r.FindByID(ctx, ctTs.ID)
}

func (r *levelDBUserRepository) FindByRecoveryToken(ctx context.Context, token string) (*user.User, error) {
	rtKey := transformer.MarshalUserKey(r.usersRecoveryTokenIndexKeyspace, token)

	rtValue, err := r.db.Get([]byte(rtKey), nil)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return nil, fmt.Errorf("%s(%s): %w", opFindByRecoveryToken, token, database.ErrNotFound)
		}

		return nil, fmt.Errorf("%s(%s): %w", opFindByRecoveryToken, token, err)
	}

	rtTs, err := transformer.UnmarshalUser(rtValue)
	if err != nil {
		return nil, fmt.Errorf("%s(%s): %w", opFindByRecoveryToken, token, err)
	}

	return r.FindByID(ctx, rtTs.ID)
}
