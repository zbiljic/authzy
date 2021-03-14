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
	"github.com/zbiljic/authzy/pkg/domain/refreshtoken"
	"github.com/zbiljic/authzy/pkg/domain/refreshtoken/storage/json/schema"
	"github.com/zbiljic/authzy/pkg/domain/refreshtoken/storage/json/transformer"
	"github.com/zbiljic/authzy/pkg/domain/refreshtoken/storage/noop"
)

const (
	refreshTokensPrefix            = "refresh_tokens"
	refreshTokensUserIDIndexPrefix = "index_refresh_tokens_user_id"
	refreshTokensTokenIndexPrefix  = "index_refresh_tokens_token"
)

// levelDBRefreshTokenRepository is a repository that uses LevelDB database.
type levelDBRefreshTokenRepository struct {
	noop.UnimplementedRefreshTokenRepository

	db *leveldb.DB
	mu sync.Mutex

	refreshTokensKeyspace            string
	refreshTokensUserIDIndexKeyspace string
	refreshTokensTokenIndexKeyspace  string

	validate *validator.Validate
}

// NewRefreshTokenRepository returns a new LevelDB repository.
func NewRefreshTokenRepository(
	db *leveldb.DB,
	keyPrefix string,
) (refreshtoken.RefreshTokenRepository, error) {
	r := &levelDBRefreshTokenRepository{
		db:                               db,
		refreshTokensKeyspace:            keyPrefix + refreshTokensPrefix,
		refreshTokensUserIDIndexKeyspace: keyPrefix + refreshTokensUserIDIndexPrefix,
		refreshTokensTokenIndexKeyspace:  keyPrefix + refreshTokensTokenIndexPrefix,
		validate:                         validator.New(),
	}

	return r, nil
}

const (
	ns               = "refreshtoken/storage/leveldb."
	opSave           = ns + "Save"
	opFindByID       = ns + "FindByID"
	opExistsByID     = ns + "ExistsByID"
	opFindAll        = ns + "FindAll"
	opCount          = ns + "Count"
	opDeleteByID     = ns + "DeleteByID"
	opDelete         = ns + "Delete"
	opFindByToken    = ns + "FindByToken"
	opFindAllForUser = ns + "FindAllForUser"
)

func (r *levelDBRefreshTokenRepository) commit(ctx context.Context, batch *leveldb.Batch) error {
	return r.db.Write(batch, nil)
}

func (r *levelDBRefreshTokenRepository) Save(ctx context.Context, entity *refreshtoken.RefreshToken) (*refreshtoken.RefreshToken, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	inS := schema.RefreshTokenToSchema(entity)

	err := r.validate.Struct(inS)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", opSave, err)
	}

	// before save
	if inS.CreatedAt.IsZero() {
		inS.CreatedAt = time.Now()
	}
	inS.UpdatedAt = time.Now()

	key := transformer.MarshalRefreshTokenKey(r.refreshTokensKeyspace, inS.ID)

	value, err := transformer.MarshalRefreshToken(inS)
	if err != nil {
		return nil, fmt.Errorf("%s(%s): %w", opSave, inS.ID, err)
	}

	batch := new(leveldb.Batch)

	batch.Put([]byte(key), value)

	uidKey := transformer.MarshalRefreshTokenUserIDKey(r.refreshTokensUserIDIndexKeyspace, inS.UserID, inS.ID)
	batch.Put([]byte(uidKey), nil)

	tKey := transformer.MarshalRefreshTokenKey(r.refreshTokensTokenIndexKeyspace, inS.Token)
	partialRefreshToken := schema.RefreshToken{
		ID:    inS.ID,
		Token: inS.Token,
	}

	partialValue, err := transformer.MarshalRefreshToken(&partialRefreshToken)
	if err != nil {
		return nil, fmt.Errorf("%s(%s): %w", opSave, inS.ID, err)
	}

	batch.Put([]byte(tKey), partialValue)

	err = r.commit(ctx, batch)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", opSave, err)
	}

	savedEntity := schema.RefreshTokenFromSchema(inS)

	return savedEntity, nil
}

func (r *levelDBRefreshTokenRepository) FindByID(ctx context.Context, id string) (*refreshtoken.RefreshToken, error) {
	key := transformer.MarshalRefreshTokenKey(r.refreshTokensKeyspace, id)

	value, err := r.db.Get([]byte(key), nil)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return nil, fmt.Errorf("%s(%s): %w", opFindByID, id, database.ErrNotFound)
		}

		return nil, fmt.Errorf("%s(%s): %w", opFindByID, id, err)
	}

	ts, err := transformer.UnmarshalRefreshToken(value)
	if err != nil {
		return nil, fmt.Errorf("%s(%s): %w", opFindByID, id, err)
	}

	entity := schema.RefreshTokenFromSchema(ts)

	return entity, nil
}

func (r *levelDBRefreshTokenRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	key := transformer.MarshalRefreshTokenKey(r.refreshTokensKeyspace, id)

	has, err := r.db.Has([]byte(key), nil)
	if err != nil {
		return false, fmt.Errorf("%s(%s): %w", opExistsByID, id, err)
	}

	return has, nil
}

func (r *levelDBRefreshTokenRepository) FindAll(ctx context.Context, afterCursor string, limit int) ([]*refreshtoken.RefreshToken, string, error) {
	if limit <= 0 {
		limit = 25
	}

	var (
		offset     int
		result     []*refreshtoken.RefreshToken
		nextCursor string
	)

	iter := r.db.NewIterator(util.BytesPrefix([]byte(r.refreshTokensKeyspace)), nil)
	defer iter.Release()

	if afterCursor != "" {
		key := transformer.MarshalRefreshTokenKey(r.refreshTokensKeyspace, afterCursor)

		if ok := iter.Seek([]byte(key)); !ok {
			err := iter.Error()
			if err != nil {
				return nil, "", fmt.Errorf("%s(%s): %w", opFindAll, afterCursor, err)
			}
		}
	}

	for iter.Next() {
		offset++

		ts, err := transformer.UnmarshalRefreshToken(iter.Value())
		if err != nil {
			return nil, "", fmt.Errorf("%s(%s): %w", opFindAll, string(iter.Key()), err)
		}

		t := schema.RefreshTokenFromSchema(ts)

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

func (r *levelDBRefreshTokenRepository) Count(ctx context.Context) (int, error) {
	var (
		count          int
		ctxCheckOffset int
	)

	iter := r.db.NewIterator(util.BytesPrefix([]byte(r.refreshTokensKeyspace)), nil)
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

func (r *levelDBRefreshTokenRepository) DeleteByID(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := transformer.MarshalRefreshTokenKey(r.refreshTokensKeyspace, id)

	value, err := r.db.Get([]byte(key), nil)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return nil
		}

		return fmt.Errorf("%s(%s): %w", opDeleteByID, id, err)
	}

	ts, err := transformer.UnmarshalRefreshToken(value)
	if err != nil {
		return fmt.Errorf("%s(%s): %w", opDeleteByID, id, err)
	}

	entity := schema.RefreshTokenFromSchema(ts)

	batch := new(leveldb.Batch)

	// delete main value
	batch.Delete([]byte(key))

	// delete from index
	uidKey := transformer.MarshalRefreshTokenUserIDKey(r.refreshTokensUserIDIndexKeyspace, entity.UserID, entity.ID)
	tKey := transformer.MarshalRefreshTokenKey(r.refreshTokensTokenIndexKeyspace, entity.Token)

	batch.Delete([]byte(uidKey))
	batch.Delete([]byte(tKey))

	err = r.commit(ctx, batch)
	if err != nil {
		return fmt.Errorf("%s: %w", opDeleteByID, err)
	}

	return nil
}

func (r *levelDBRefreshTokenRepository) Delete(ctx context.Context, entity *refreshtoken.RefreshToken) error {
	inS := schema.RefreshTokenToSchema(entity)

	err := r.validate.Struct(inS)
	if err != nil {
		return fmt.Errorf("%s: %w", opDelete, err)
	}

	return r.DeleteByID(ctx, inS.ID)
}

func (r *levelDBRefreshTokenRepository) FindByToken(ctx context.Context, token string) (*refreshtoken.RefreshToken, error) {
	tKey := transformer.MarshalRefreshTokenKey(r.refreshTokensTokenIndexKeyspace, token)

	tValue, err := r.db.Get([]byte(tKey), nil)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return nil, fmt.Errorf("%s(%s): %w", opFindByToken, token, database.ErrNotFound)
		}

		return nil, fmt.Errorf("%s(%s): %w", opFindByToken, token, err)
	}

	ts, err := transformer.UnmarshalRefreshToken(tValue)
	if err != nil {
		return nil, fmt.Errorf("%s(%s): %w", opFindByToken, token, err)
	}

	return r.FindByID(ctx, ts.ID)
}

func (r *levelDBRefreshTokenRepository) FindAllForUser(ctx context.Context, userID, afterCursor string, limit int) ([]*refreshtoken.RefreshToken, string, error) {
	if userID == "" {
		return nil, "", fmt.Errorf("%s: userID cannot be empty", opFindAllForUser)
	}

	if limit <= 0 {
		limit = 25
	}

	var (
		offset     int
		result     []*refreshtoken.RefreshToken
		nextCursor string
	)

	keyPrefix := transformer.MarshalRefreshTokenUserIDKey(r.refreshTokensUserIDIndexKeyspace, userID, "")

	iter := r.db.NewIterator(util.BytesPrefix([]byte(keyPrefix)), nil)
	defer iter.Release()

	if afterCursor != "" {
		key := transformer.MarshalRefreshTokenUserIDKey(r.refreshTokensUserIDIndexKeyspace, userID, afterCursor)

		if ok := iter.Seek([]byte(key)); !ok {
			err := iter.Error()
			if err != nil {
				return nil, "", fmt.Errorf("%s(%s): %w", opFindAllForUser, afterCursor, err)
			}
		}
	}

	for iter.Next() {
		offset++

		_, kID := transformer.UnmarshalRefreshTokenUserIDKey(string(iter.Key()))

		t, err := r.FindByID(ctx, kID)
		if err != nil {
			// ignore
			continue
		}

		result = append(result, t)

		if limit == offset {
			break // stops iterator
		}
	}

	err := iter.Error()
	if err != nil {
		return nil, "", fmt.Errorf("%s(%s): %w", opFindAllForUser, userID, err)
	}

	if len(result) == limit {
		nextCursor = result[len(result)-1].ID
	}

	return result, nextCursor, nil
}
