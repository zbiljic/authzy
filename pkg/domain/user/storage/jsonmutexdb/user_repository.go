package jsonmutexdb

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/zbiljic/authzy/pkg/database"
	"github.com/zbiljic/authzy/pkg/database/jsonmutexdb"
	"github.com/zbiljic/authzy/pkg/domain/user"
	"github.com/zbiljic/authzy/pkg/domain/user/storage/json/schema"
	"github.com/zbiljic/authzy/pkg/domain/user/storage/noop"
)

const (
	usersPrefix = "users"
)

type jsonMutexDBUserRepository struct {
	noop.UnimplementedUserRepository

	db                            map[string]schema.User
	dbIndexUsersIdentifier        map[string]*schema.User
	dbIndexUsersConfirmationToken map[string]*schema.User
	dbIndexUsersRecoveryToken     map[string]*schema.User
	mu                            sync.RWMutex

	loadSaver jsonmutexdb.LoadSaver
	filename  string

	validate *validator.Validate
}

// NewUserRepository returns a new JSONMutexDB repository.
func NewUserRepository(
	loadSaver jsonmutexdb.LoadSaver,
	filenamePrefix string,
) (user.UserRepository, error) {
	validate := validator.New()
	schema.RegisterValidators(validate)

	r := &jsonMutexDBUserRepository{
		db:                            make(map[string]schema.User),
		dbIndexUsersIdentifier:        make(map[string]*schema.User),
		dbIndexUsersConfirmationToken: make(map[string]*schema.User),
		dbIndexUsersRecoveryToken:     make(map[string]*schema.User),
		loadSaver:                     loadSaver,
		filename:                      fmt.Sprintf("%s%s.json", filenamePrefix, usersPrefix),
		validate:                      validate,
	}

	if err := r.load(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *jsonMutexDBUserRepository) load() error {
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

			// read indexes
			for _, v := range r.db {
				v := v
				r.dbIndexUsersIdentifier[v.NormalizedUsername] = &v
				r.dbIndexUsersIdentifier[v.Email] = &v
				if v.ConfirmationToken != "" {
					r.dbIndexUsersConfirmationToken[v.ConfirmationToken] = &v
				}
				if v.RecoveryToken != "" {
					r.dbIndexUsersRecoveryToken[v.RecoveryToken] = &v
				}
			}
		}
	}
	return nil
}

func (r *jsonMutexDBUserRepository) commit(ctx context.Context) error {
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
	ns                        = "user/storage/jsonmutexdb."
	opSave                    = ns + "Save"
	opFindByID                = ns + "FindByID"
	opDeleteByID              = ns + "DeleteByID"
	opFindByIdentifier        = ns + "FindByIdentifier"
	opFindByConfirmationToken = ns + "FindByConfirmationToken"
	opFindByRecoveryToken     = ns + "FindByRecoveryToken"
)

func (r *jsonMutexDBUserRepository) Save(ctx context.Context, entity *user.User) (*user.User, error) {
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

	r.db[inS.ID] = *inS

	if inS.PasswordHash != "" {
		r.dbIndexUsersIdentifier[inS.NormalizedUsername] = inS
		r.dbIndexUsersIdentifier[inS.Email] = inS
	}

	if inS.ConfirmationToken != "" {
		r.dbIndexUsersConfirmationToken[inS.ConfirmationToken] = inS
	} else {
		_, has := r.db[inS.ConfirmationToken]

		if has {
			delete(r.db, inS.ConfirmationToken)
		}
	}

	if inS.RecoveryToken != "" {
		r.dbIndexUsersRecoveryToken[inS.RecoveryToken] = inS
	} else {
		_, has := r.db[inS.RecoveryToken]

		if has {
			delete(r.db, inS.RecoveryToken)
		}
	}

	err = r.commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", opSave, err)
	}

	savedEntity := schema.UserFromSchema(inS)

	return savedEntity, nil
}

func (r *jsonMutexDBUserRepository) FindByID(ctx context.Context, id string) (*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	value, ok := r.db[id]
	if !ok {
		return nil, fmt.Errorf("%s(%s): %w", opFindByID, id, database.ErrNotFound)
	}

	entity := schema.UserFromSchema(&value)

	return entity, nil
}

func (r *jsonMutexDBUserRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, has := r.db[id]

	return has, nil
}

func (r *jsonMutexDBUserRepository) FindAll(ctx context.Context, afterCursor string, limit int) ([]*user.User, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if limit <= 0 {
		limit = 25
	}

	var (
		offset     int
		result     []*user.User
		nextCursor string
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

		t := schema.UserFromSchema(&val)

		result = append(result, t)

		if limit == offset {
			break // stops iterator
		}
	}

	if len(result) == limit {
		nextCursor = result[len(result)-1].ID
	}

	return result, nextCursor, nil
}

func (r *jsonMutexDBUserRepository) Count(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.db), nil
}

func (r *jsonMutexDBUserRepository) DeleteByID(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	value, ok := r.db[id]
	if !ok {
		return nil
	}

	entity := schema.UserFromSchema(&value)

	// delete main value
	delete(r.db, id)

	// delete from index
	delete(r.dbIndexUsersIdentifier, entity.NormalizedUsername)
	delete(r.dbIndexUsersIdentifier, entity.Email)

	if entity.ConfirmationToken != "" {
		delete(r.dbIndexUsersConfirmationToken, entity.ConfirmationToken)
	}

	if entity.RecoveryToken != "" {
		delete(r.dbIndexUsersRecoveryToken, entity.RecoveryToken)
	}

	err := r.commit(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", opDeleteByID, err)
	}

	return nil
}

func (r *jsonMutexDBUserRepository) Delete(ctx context.Context, entity *user.User) error {
	return r.DeleteByID(ctx, entity.ID)
}

func (r *jsonMutexDBUserRepository) DeleteAll(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.db = make(map[string]schema.User)
	r.dbIndexUsersIdentifier = make(map[string]*schema.User)
	r.dbIndexUsersConfirmationToken = make(map[string]*schema.User)
	r.dbIndexUsersRecoveryToken = make(map[string]*schema.User)

	return nil
}

func (r *jsonMutexDBUserRepository) ExistsByIdentifier(ctx context.Context, identifier string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, has := r.dbIndexUsersIdentifier[identifier]

	return has, nil
}

func (r *jsonMutexDBUserRepository) FindByIdentifier(ctx context.Context, identifier string) (*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// find using index first
	identifierValue, ok := r.dbIndexUsersIdentifier[identifier]
	if !ok {
		return nil, fmt.Errorf("%s(%s): %w", opFindByIdentifier, identifier, database.ErrNotFound)
	}

	return r.FindByID(ctx, identifierValue.ID)
}

func (r *jsonMutexDBUserRepository) FindByConfirmationToken(ctx context.Context, token string) (*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ctValue, ok := r.dbIndexUsersConfirmationToken[token]
	if !ok {
		return nil, fmt.Errorf("%s(%s): %w", opFindByConfirmationToken, token, database.ErrNotFound)
	}

	return r.FindByID(ctx, ctValue.ID)
}

func (r *jsonMutexDBUserRepository) FindByRecoveryToken(ctx context.Context, token string) (*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rtValue, ok := r.dbIndexUsersRecoveryToken[token]
	if !ok {
		return nil, fmt.Errorf("%s(%s): %w", opFindByRecoveryToken, token, database.ErrNotFound)
	}

	return r.FindByID(ctx, rtValue.ID)
}
