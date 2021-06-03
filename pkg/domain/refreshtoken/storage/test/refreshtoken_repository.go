package test

import (
	"context"
	"errors"
	"sort"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zbiljic/authzy/pkg/database"
	"github.com/zbiljic/authzy/pkg/domain/refreshtoken"
	"github.com/zbiljic/authzy/pkg/domain/refreshtoken/storage/json/schema"
	"github.com/zbiljic/authzy/pkg/ulid"
)

const (
	TestPrefix = "test-"
)

func testValidator() *validator.Validate {
	return validator.New()
}

func createRefreshTokens(t *testing.T, repo refreshtoken.RefreshTokenRepository, userID string, count int) []*refreshtoken.RefreshToken {
	t.Helper()

	var result []*refreshtoken.RefreshToken

	ctx := context.Background()

	for i := 0; i < count; i++ {
		entity := &refreshtoken.RefreshToken{
			ID:     ulid.ULID().String(),
			UserID: userID,
			Token:  ulid.ULID().String(),
		}

		savedEntity, err := repo.Save(ctx, entity)
		assert.NoError(t, err)

		result = append(result, savedEntity)
	}

	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })

	return result
}

func assertRefreshTokenEqual(t *testing.T, expected, actual *refreshtoken.RefreshToken) {
	t.Helper()

	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.UserID, actual.UserID)
	assert.Equal(t, expected.Token, actual.Token)
	assert.Equal(t, expected.Revoked, actual.Revoked)
	assert.Equal(t, expected.CreatedAt.Unix(), actual.CreatedAt.Unix())
	assert.Equal(t, expected.UpdatedAt.Unix(), actual.UpdatedAt.Unix())
}

func Run(t *testing.T, f func() func(t *testing.T) (refreshtoken.RefreshTokenRepository, func())) {
	t.Helper()

	t.Run("init", func(t *testing.T) {
		_, cleanup := f()(t)
		defer cleanup()
	})
	t.Run("Save", func(t *testing.T) {
		repo, cleanup := f()(t)
		defer cleanup()

		testRefreshTokenRepositorySave(t, repo)
	})
	t.Run("FindByID", func(t *testing.T) {
		repo, cleanup := f()(t)
		defer cleanup()

		testRefreshTokenRepositoryFindByID(t, repo)
	})
	t.Run("ExistsByID", func(t *testing.T) {
		repo, cleanup := f()(t)
		defer cleanup()

		testRefreshTokenRepositoryExistsByID(t, repo)
	})
	t.Run("FindAll", func(t *testing.T) {
		repo, cleanup := f()(t)
		defer cleanup()

		testRefreshTokenRepositoryFindAll(t, repo)
	})
	t.Run("Count", func(t *testing.T) {
		repo, cleanup := f()(t)
		defer cleanup()

		testRefreshTokenRepositoryCount(t, repo)
	})
	t.Run("DeleteByID", func(t *testing.T) {
		repo, cleanup := f()(t)
		defer cleanup()

		testRefreshTokenRepositoryDeleteByID(t, repo)
	})
	t.Run("FindByToken", func(t *testing.T) {
		repo, cleanup := f()(t)
		defer cleanup()

		testRefreshTokenRepositoryFindByToken(t, repo)
	})
	t.Run("FindAllForUser", func(t *testing.T) {
		repo, cleanup := f()(t)
		defer cleanup()

		testRefreshTokenRepositoryFindAllForUser(t, repo)
	})
}

func testRefreshTokenRepositorySave(t *testing.T, repo refreshtoken.RefreshTokenRepository) {
	t.Helper()

	ctx := context.Background()
	validate := testValidator()

	t.Run("nil", func(t *testing.T) {
		_, err := repo.Save(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("empty", func(t *testing.T) {
		entity := &refreshtoken.RefreshToken{}

		_, err := repo.Save(ctx, entity)
		assert.Error(t, err)

		inS := schema.RefreshTokenToSchema(entity)
		validateErr := validate.Struct(inS)

		assert.Contains(t, err.Error(), validateErr.Error())
	})

	t.Run("simple", func(t *testing.T) {
		entity := &refreshtoken.RefreshToken{
			ID:      "0",
			UserID:  "0",
			Token:   "0",
			Revoked: false,
		}

		_, err := repo.Save(ctx, entity)
		assert.NoError(t, err)
	})
}

func testRefreshTokenRepositoryFindByID(t *testing.T, repo refreshtoken.RefreshTokenRepository) {
	t.Helper()

	ctx := context.Background()
	userID := "test"

	tokens := createRefreshTokens(t, repo, userID, 1)

	t.Run("non existent", func(t *testing.T) {
		_, err := repo.FindByID(ctx, "non_existent_id")
		require.Error(t, err)

		assert.True(t, errors.Is(err, database.ErrNotFound))
	})

	t.Run("ok", func(t *testing.T) {
		entity, err := repo.FindByID(ctx, tokens[0].ID)
		require.NoError(t, err)

		assert.NotNil(t, entity)
		assertRefreshTokenEqual(t, tokens[0], entity)
	})
}

func testRefreshTokenRepositoryExistsByID(t *testing.T, repo refreshtoken.RefreshTokenRepository) {
	t.Helper()

	ctx := context.Background()
	userID := "test"

	tokens := createRefreshTokens(t, repo, userID, 1)

	t.Run("non existent", func(t *testing.T) {
		exists, err := repo.ExistsByID(ctx, "non_existent_id")
		require.NoError(t, err)

		assert.False(t, exists)
	})

	t.Run("ok", func(t *testing.T) {
		exists, err := repo.ExistsByID(ctx, tokens[0].ID)
		require.NoError(t, err)

		assert.True(t, exists)
	})
}

func testRefreshTokenRepositoryFindAll(t *testing.T, repo refreshtoken.RefreshTokenRepository) {
	t.Helper()

	ctx := context.Background()
	userID := "test"

	t.Run("empty", func(t *testing.T) {
		results, nextCursor, err := repo.FindAll(ctx, "", 0)
		assert.NoError(t, err)

		assert.Equal(t, 0, len(results))
		assert.Equal(t, "", nextCursor)
	})

	createCount := 7

	tokens := createRefreshTokens(t, repo, userID, createCount)

	t.Run("ok", func(t *testing.T) {
		results, nextCursor, err := repo.FindAll(ctx, "", 0)
		assert.NoError(t, err)

		assert.Equal(t, createCount, len(results))
		assert.Equal(t, "", nextCursor)
	})

	t.Run("paging", func(t *testing.T) {
		limit := 5

		results, nextCursor, err := repo.FindAll(ctx, "", limit)
		assert.NoError(t, err)

		assert.Equal(t, limit, len(results))
		assert.Equal(t, tokens[limit-1].ID, nextCursor)

		// next page
		results, nextCursor, err = repo.FindAll(ctx, nextCursor, limit)
		assert.NoError(t, err)

		assert.Equal(t, createCount-limit, len(results))
		assert.Equal(t, "", nextCursor)
	})
}

func testRefreshTokenRepositoryCount(t *testing.T, repo refreshtoken.RefreshTokenRepository) {
	t.Helper()

	ctx := context.Background()
	userID := "test"

	count, err := repo.Count(ctx)
	require.NoError(t, err)

	assert.Equal(t, 0, count)

	createCount := 3

	createRefreshTokens(t, repo, userID, createCount)

	count, err = repo.Count(ctx)
	require.NoError(t, err)

	assert.Equal(t, createCount, count)
}

func testRefreshTokenRepositoryDeleteByID(t *testing.T, repo refreshtoken.RefreshTokenRepository) {
	t.Helper()

	ctx := context.Background()
	userID := "test"

	tokens := createRefreshTokens(t, repo, userID, 1)

	exists, err := repo.ExistsByID(ctx, tokens[0].ID)
	require.NoError(t, err)

	assert.True(t, exists)

	t.Run("non existent", func(t *testing.T) {
		err := repo.DeleteByID(ctx, "non_existent_id")
		require.NoError(t, err)

		exists, err = repo.ExistsByID(ctx, tokens[0].ID)
		require.NoError(t, err)

		assert.True(t, exists)
	})

	t.Run("ok", func(t *testing.T) {
		err := repo.DeleteByID(ctx, tokens[0].ID)
		require.NoError(t, err)

		exists, err = repo.ExistsByID(ctx, tokens[0].ID)
		require.NoError(t, err)

		assert.False(t, exists)
	})
}

func testRefreshTokenRepositoryFindByToken(t *testing.T, repo refreshtoken.RefreshTokenRepository) {
	t.Helper()

	ctx := context.Background()
	userID := "test"

	tokens := createRefreshTokens(t, repo, userID, 1)

	t.Run("non existent", func(t *testing.T) {
		_, err := repo.FindByToken(ctx, "non_existent_id")
		require.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		entity, err := repo.FindByToken(ctx, tokens[0].Token)
		require.NoError(t, err)

		assert.NotNil(t, entity)
		assertRefreshTokenEqual(t, tokens[0], entity)
	})
}

func testRefreshTokenRepositoryFindAllForUser(t *testing.T, repo refreshtoken.RefreshTokenRepository) {
	t.Helper()

	ctx := context.Background()
	user1ID := "test1"
	user2ID := "test2"

	t.Run("empty", func(t *testing.T) {
		results, _, err := repo.FindAllForUser(ctx, "", "", 0)
		assert.Error(t, err)

		assert.Equal(t, 0, len(results))
	})

	createCountUser1 := 2
	createCountUser2 := 1

	var tokens []*refreshtoken.RefreshToken

	tokens = append(tokens, createRefreshTokens(t, repo, user1ID, createCountUser1)...)
	tokens = append(tokens, createRefreshTokens(t, repo, user2ID, createCountUser2)...)

	t.Run("ok", func(t *testing.T) {
		results1, _, err := repo.FindAllForUser(ctx, user1ID, "", 0)
		assert.NoError(t, err)
		assert.Equal(t, createCountUser1, len(results1))

		results2, _, err := repo.FindAllForUser(ctx, user2ID, "", 0)
		assert.NoError(t, err)
		assert.Equal(t, createCountUser2, len(results2))

		assert.Equal(t, len(tokens), len(results1)+len(results2))
	})
}
