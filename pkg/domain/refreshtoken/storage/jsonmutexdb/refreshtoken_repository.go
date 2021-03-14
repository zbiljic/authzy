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
	"github.com/zbiljic/authzy/pkg/domain/refreshtoken"
	"github.com/zbiljic/authzy/pkg/domain/refreshtoken/storage/json/schema"
	"github.com/zbiljic/authzy/pkg/domain/refreshtoken/storage/noop"
)

const (
	refreshTokensPrefix = "refresh_tokens"
)

type jsonMutexDBRefreshTokenRepository struct {
	noop.UnimplementedRefreshTokenRepository

	db           map[string]schema.RefreshToken
	dbIndexToken map[string]*schema.RefreshToken
	mu           sync.RWMutex

	loadSaver jsonmutexdb.LoadSaver
	filename  string

	validate *validator.Validate
}

// NewRefreshTokenRepository returns a new JSONMutexDB repository.
func NewRefreshTokenRepository(
	loadSaver jsonmutexdb.LoadSaver,
	filenamePrefix string,
) (refreshtoken.RefreshTokenRepository, error) {
	r := &jsonMutexDBRefreshTokenRepository{
		db:           make(map[string]schema.RefreshToken),
		dbIndexToken: make(map[string]*schema.RefreshToken),
		loadSaver:    loadSaver,
		filename:     fmt.Sprintf("%s%s.json", filenamePrefix, refreshTokensPrefix),
		validate:     validator.New(),
	}

	if err := r.load(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *jsonMutexDBRefreshTokenRepository) load() error {
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
				r.dbIndexToken[v.Token] = &v
			}
		}
	}

	return nil
}

func (r *jsonMutexDBRefreshTokenRepository) commit(ctx context.Context) error {
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
	ns               = "refreshtoken/storage/jsonmutexdb."
	opSave           = ns + "Save"
	opFindByID       = ns + "FindByID"
	opDeleteByID     = ns + "DeleteByID"
	opFindByToken    = ns + "FindByToken"
	opFindAllForUser = ns + "FindAllForUser"
)

func (r *jsonMutexDBRefreshTokenRepository) Save(ctx context.Context, entity *refreshtoken.RefreshToken) (*refreshtoken.RefreshToken, error) {
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

	r.db[inS.ID] = *inS

	r.dbIndexToken[inS.Token] = inS

	err = r.commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", opSave, err)
	}

	savedEntity := schema.RefreshTokenFromSchema(inS)

	return savedEntity, nil
}

func (r *jsonMutexDBRefreshTokenRepository) FindByID(ctx context.Context, id string) (*refreshtoken.RefreshToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	value, ok := r.db[id]
	if !ok {
		return nil, fmt.Errorf("%s(%s): %w", opFindByID, id, database.ErrNotFound)
	}

	entity := schema.RefreshTokenFromSchema(&value)

	return entity, nil
}

func (r *jsonMutexDBRefreshTokenRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, has := r.db[id]

	return has, nil
}

func (r *jsonMutexDBRefreshTokenRepository) FindAll(ctx context.Context, afterCursor string, limit int) ([]*refreshtoken.RefreshToken, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if limit <= 0 {
		limit = 25
	}

	var (
		offset     int
		result     []*refreshtoken.RefreshToken
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

		t := schema.RefreshTokenFromSchema(&val)

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

func (r *jsonMutexDBRefreshTokenRepository) Count(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.db), nil
}

func (r *jsonMutexDBRefreshTokenRepository) DeleteByID(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	value, ok := r.db[id]
	if !ok {
		return nil
	}

	entity := schema.RefreshTokenFromSchema(&value)

	// delete main value
	delete(r.db, entity.ID)

	// delete from index
	delete(r.dbIndexToken, entity.Token)

	err := r.commit(ctx)
	if err != nil {
		return fmt.Errorf("%s(%s): %w", opDeleteByID, id, err)
	}

	return nil
}

func (r *jsonMutexDBRefreshTokenRepository) Delete(ctx context.Context, entity *refreshtoken.RefreshToken) error {
	return r.DeleteByID(ctx, entity.ID)
}

func (r *jsonMutexDBRefreshTokenRepository) DeleteAll(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.db = make(map[string]schema.RefreshToken)
	r.dbIndexToken = make(map[string]*schema.RefreshToken)

	return nil
}

func (r *jsonMutexDBRefreshTokenRepository) FindByToken(ctx context.Context, token string) (*refreshtoken.RefreshToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tValue, ok := r.dbIndexToken[token]
	if !ok {
		return nil, fmt.Errorf("%s(%s): %w", opFindByToken, token, database.ErrNotFound)
	}

	return r.FindByID(ctx, tValue.ID)
}

func (r *jsonMutexDBRefreshTokenRepository) FindAllForUser(ctx context.Context, userID, afterCursor string, limit int) ([]*refreshtoken.RefreshToken, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

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
		val := r.db[id]

		if userID != val.UserID {
			continue
		}

		if afterCursorKey != "" {
			if afterCursorKey == id {
				afterCursorKey = ""
			}

			continue
		}

		offset++

		t := schema.RefreshTokenFromSchema(&val)

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
