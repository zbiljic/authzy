package jsonmutexdb

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"github.com/go-playground/validator/v10"

	"github.com/zbiljic/authzy/pkg/database"
	"github.com/zbiljic/authzy/pkg/database/jsonmutexdb"
	"github.com/zbiljic/authzy/pkg/domain/account"
	"github.com/zbiljic/authzy/pkg/domain/account/storage/json/schema"
	"github.com/zbiljic/authzy/pkg/domain/account/storage/json/transformer"
	"github.com/zbiljic/authzy/pkg/domain/account/storage/noop"
)

const (
	accountsPrefix = "accounts"
)

type jsonMutexDBAccountRepository struct {
	noop.UnimplementedAccountRepository

	db map[string]schema.Account
	mu sync.RWMutex

	loadSaver jsonmutexdb.LoadSaver
	filename  string

	validate *validator.Validate
}

// NewAccountRepository returns a new JSONMutexDB repository.
func NewAccountRepository(
	loadSaver jsonmutexdb.LoadSaver,
	filenamePrefix string,
) (account.AccountRepository, error) {
	r := &jsonMutexDBAccountRepository{
		db:        make(map[string]schema.Account),
		loadSaver: loadSaver,
		filename:  fmt.Sprintf("%s%s.json", filenamePrefix, accountsPrefix),
		validate:  validator.New(),
	}

	if err := r.load(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *jsonMutexDBAccountRepository) load() error {
	if r.loadSaver != nil {
		data, err := r.loadSaver.Load(r.filename)
		if err != nil {
			return err
		}

		if len(data) > 0 {
			err = json.Unmarshal(data, &r.db)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *jsonMutexDBAccountRepository) commit(ctx context.Context) error {
	if r.loadSaver != nil {
		out, err := json.Marshal(r.db)
		if err != nil {
			return err
		}

		return r.loadSaver.Save(r.filename, out)
	}
	return nil
}

const (
	ns               = "account/storage/jsonmutexdb."
	opSave           = ns + "Save"
	opFind           = ns + "Find"
	opExists         = ns + "Exists"
	opDelete         = ns + "Delete"
	opFindAllForUser = ns + "FindAllForUser"
)

func (r *jsonMutexDBAccountRepository) Save(ctx context.Context, entity *account.Account) (*account.Account, error) {
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

	r.db[id] = *inS

	err = r.commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", opSave, err)
	}

	savedEntity := schema.AccountFromSchema(inS)

	return savedEntity, nil
}

func (r *jsonMutexDBAccountRepository) Find(ctx context.Context, entity *account.Account) (*account.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	inS := schema.AccountToSchema(entity)

	err := r.validate.Struct(inS)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", opFind, err)
	}

	id := transformer.MarshalAccountID(inS)

	value, ok := r.db[id]
	if !ok {
		return nil, fmt.Errorf("%s(%s): %w", opFind, id, database.ErrNotFound)
	}

	dbEntity := schema.AccountFromSchema(&value)

	return dbEntity, nil
}

func (r *jsonMutexDBAccountRepository) Exists(ctx context.Context, entity *account.Account) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	inS := schema.AccountToSchema(entity)

	err := r.validate.Struct(inS)
	if err != nil {
		return false, fmt.Errorf("%s: %w", opExists, err)
	}

	id := transformer.MarshalAccountID(inS)

	_, has := r.db[id]

	return has, nil
}

func (r *jsonMutexDBAccountRepository) FindAll(ctx context.Context, afterCursor string, limit int) ([]*account.Account, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if limit <= 0 {
		limit = 25
	}

	var (
		offset       int
		result       []*account.Account
		lastResultID string
		nextCursor   string
	)

	keys := []string{}
	for id := range r.db {
		keys = append(keys, id)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	var afterCursorKey string
	if afterCursor != "" {
		afterCursorKey = afterCursor
	}

	for _, id := range keys {
		if afterCursorKey != "" {
			if afterCursorKey == id {
				afterCursorKey = ""
			}

			continue
		}

		offset++

		val := r.db[id]

		lastResultID = transformer.MarshalAccountID(&val)

		t := schema.AccountFromSchema(&val)

		result = append(result, t)

		if limit == offset {
			break // stops iterator
		}
	}

	if len(result) == limit {
		nextCursor = lastResultID
	}

	return result, nextCursor, nil
}

func (r *jsonMutexDBAccountRepository) Count(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.db), nil
}

func (r *jsonMutexDBAccountRepository) Delete(ctx context.Context, entity *account.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	inS := schema.AccountToSchema(entity)

	err := r.validate.Struct(inS)
	if err != nil {
		return fmt.Errorf("%s: %w", opDelete, err)
	}

	id := transformer.MarshalAccountID(inS)

	delete(r.db, id)

	return nil
}

func (r *jsonMutexDBAccountRepository) DeleteAll(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.db = make(map[string]schema.Account)

	return nil
}

func (r *jsonMutexDBAccountRepository) FindAllForUser(ctx context.Context, userID string) ([]*account.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if userID == "" {
		return nil, fmt.Errorf("%s: userID cannot be empty", opFindAllForUser)
	}

	var (
		result []*account.Account
	)

	for _, val := range r.db {
		if val.UserID != userID {
			continue
		}

		val := val
		t := schema.AccountFromSchema(&val)

		result = append(result, t)
	}

	return result, nil
}
