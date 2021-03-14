package leveldb

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/zbiljic/authzy/pkg/database"
	"github.com/zbiljic/authzy/pkg/domain/account"
	"github.com/zbiljic/authzy/pkg/domain/account/storage/json/schema"
	"github.com/zbiljic/authzy/pkg/domain/account/storage/json/transformer"
	"github.com/zbiljic/authzy/pkg/domain/account/storage/noop"
)

const (
	accountsPrefix = "accounts"
)

// levelDBAccountRepository is a repository that uses LevelDB database.
type levelDBAccountRepository struct {
	noop.UnimplementedAccountRepository

	db *leveldb.DB
	mu sync.Mutex

	accountsKeyspace string

	validate *validator.Validate
}

// NewAccountRepository returns a new LevelDB repository.
func NewAccountRepository(
	db *leveldb.DB,
	keyPrefix string,
) (account.AccountRepository, error) {
	r := &levelDBAccountRepository{
		db:               db,
		accountsKeyspace: keyPrefix + accountsPrefix,
		validate:         validator.New(),
	}

	return r, nil
}

const (
	ns               = "account/storage/leveldb."
	opSave           = ns + "Save"
	opFind           = ns + "Find"
	opExists         = ns + "Exists"
	opFindAll        = ns + "FindAll"
	opCount          = ns + "Count"
	opDelete         = ns + "Delete"
	opFindAllForUser = ns + "FindAllForUser"
)

func (r *levelDBAccountRepository) Save(ctx context.Context, entity *account.Account) (*account.Account, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	inS := schema.AccountToSchema(entity)

	err := r.validate.Struct(inS)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", opSave, err)
	}

	// before save
	err = inS.BeforeSave()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", opSave, err)
	}

	id := transformer.MarshalAccountID(inS)
	key := transformer.MarshalAccountKey(r.accountsKeyspace, id)

	value, err := transformer.MarshalAccount(inS)
	if err != nil {
		return nil, fmt.Errorf("%s(%s): %w", opSave, id, err)
	}

	err = r.db.Put([]byte(key), value, nil)
	if err != nil {
		return nil, fmt.Errorf("%s(%s): %w", opSave, id, err)
	}

	savedEntity := schema.AccountFromSchema(inS)

	return savedEntity, nil
}

func (r *levelDBAccountRepository) Find(ctx context.Context, entity *account.Account) (*account.Account, error) {
	inS := schema.AccountToSchema(entity)

	err := r.validate.Struct(inS)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", opFind, err)
	}

	id := transformer.MarshalAccountID(inS)
	key := transformer.MarshalAccountKey(r.accountsKeyspace, id)

	value, err := r.db.Get([]byte(key), nil)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return nil, fmt.Errorf("%s(%s): %w", opFind, id, database.ErrNotFound)
		}

		return nil, fmt.Errorf("%s(%s): %w", opFind, id, err)
	}

	ts, err := transformer.UnmarshalAccount(value)
	if err != nil {
		return nil, fmt.Errorf("%s(%s): %w", opFind, id, err)
	}

	dbEntity := schema.AccountFromSchema(ts)

	return dbEntity, nil
}

func (r *levelDBAccountRepository) Exists(ctx context.Context, entity *account.Account) (bool, error) {
	inS := schema.AccountToSchema(entity)

	err := r.validate.Struct(inS)
	if err != nil {
		return false, fmt.Errorf("%s: %w", opExists, err)
	}

	id := transformer.MarshalAccountID(inS)
	key := transformer.MarshalAccountKey(r.accountsKeyspace, id)

	has, err := r.db.Has([]byte(key), nil)
	if err != nil {
		return false, fmt.Errorf("%s(%s): %w", opExists, id, err)
	}

	return has, nil
}

func (r *levelDBAccountRepository) FindAll(ctx context.Context, afterCursor string, limit int) ([]*account.Account, string, error) {
	if limit <= 0 {
		limit = 25
	}

	var (
		offset       int
		result       []*account.Account
		lastResultID string
		nextCursor   string
	)

	iter := r.db.NewIterator(util.BytesPrefix([]byte(r.accountsKeyspace)), nil)
	defer iter.Release()

	if afterCursor != "" {
		key := transformer.MarshalAccountKey(r.accountsKeyspace, afterCursor)

		if ok := iter.Seek([]byte(key)); !ok {
			err := iter.Error()
			if err != nil {
				return nil, "", fmt.Errorf("%s(%s): %w", opFindAll, afterCursor, err)
			}
		}
	}

	for iter.Next() {
		offset++

		ts, err := transformer.UnmarshalAccount(iter.Value())
		if err != nil {
			return nil, "", fmt.Errorf("%s(%s): %w", opFindAll, string(iter.Key()), err)
		}

		lastResultID = transformer.MarshalAccountID(ts)

		t := schema.AccountFromSchema(ts)

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
		nextCursor = lastResultID
	}

	return result, nextCursor, nil
}

func (r *levelDBAccountRepository) Count(ctx context.Context) (int, error) {
	var (
		count          int
		ctxCheckOffset int
	)

	iter := r.db.NewIterator(util.BytesPrefix([]byte(r.accountsKeyspace)), nil)
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

func (r *levelDBAccountRepository) Delete(ctx context.Context, entity *account.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	inS := schema.AccountToSchema(entity)

	err := r.validate.Struct(inS)
	if err != nil {
		return fmt.Errorf("%s: %w", opDelete, err)
	}

	id := transformer.MarshalAccountID(inS)
	key := transformer.MarshalAccountKey(r.accountsKeyspace, id)

	has, err := r.db.Has([]byte(key), nil)
	if err != nil {
		return fmt.Errorf("%s(%s): %w", opDelete, key, err)
	}

	if !has {
		return nil
	}

	err = r.db.Delete([]byte(key), nil)
	if err != nil {
		return fmt.Errorf("%s: %w", opDelete, err)
	}

	return nil
}

func (r *levelDBAccountRepository) FindAllForUser(ctx context.Context, userID string) ([]*account.Account, error) {
	if userID == "" {
		return nil, fmt.Errorf("%s: userID cannot be empty", opFindAllForUser)
	}

	var (
		result []*account.Account
	)

	keyPrefix := transformer.MarshalAccountKey(r.accountsKeyspace, userID)

	iter := r.db.NewIterator(util.BytesPrefix([]byte(keyPrefix)), nil)
	defer iter.Release()

	for iter.Next() {
		ts, err := transformer.UnmarshalAccount(iter.Value())
		if err != nil {
			return nil, fmt.Errorf("%s(%s): %w", opFindAllForUser, string(iter.Key()), err)
		}

		t := schema.AccountFromSchema(ts)

		result = append(result, t)
	}

	err := iter.Error()
	if err != nil {
		return nil, fmt.Errorf("%s(%s): %w", opFindAllForUser, userID, err)
	}

	return result, nil
}
