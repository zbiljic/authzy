package test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zbiljic/authzy/pkg/database"
	"github.com/zbiljic/authzy/pkg/domain/user"
	"github.com/zbiljic/authzy/pkg/domain/user/storage/json/schema"
)

const (
	TestPrefix = "test-"
)

func testValidator() *validator.Validate {
	validate := validator.New()
	schema.RegisterValidators(validate)

	return validate
}

func createUsers(t *testing.T, repo user.UserRepository, count int) []*user.User {
	t.Helper()

	var result []*user.User

	ctx := context.Background()

	i := 0
	for i < count {
		i++

		entity := &user.User{
			ID:                 strconv.FormatInt(int64(i), 10),
			Email:              fmt.Sprintf("user_%d@test.com", i),
			PasswordHash:       fmt.Sprintf("password_%d", i),
			Username:           fmt.Sprintf("username_%d", i),
			NormalizedUsername: fmt.Sprintf("username_%d", i),
		}

		savedEntity, err := repo.Save(ctx, entity)
		assert.NoError(t, err)

		result = append(result, savedEntity)
	}

	return result
}

func assertUserEqual(t *testing.T, expected, actual *user.User) {
	t.Helper()

	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Email, actual.Email)
	assert.Equal(t, expected.PasswordHash, actual.PasswordHash)
	assert.Equal(t, expected.Username, actual.Username)
	assert.Equal(t, expected.NormalizedUsername, actual.NormalizedUsername)
	assert.Equal(t, expected.CreatedAt.Unix(), actual.CreatedAt.Unix())
	assert.Equal(t, expected.UpdatedAt.Unix(), actual.UpdatedAt.Unix())
}

func Run(t *testing.T, f func() func(t *testing.T) (user.UserRepository, func())) {
	t.Helper()

	t.Run("init", func(t *testing.T) {
		_, cleanup := f()(t)
		defer cleanup()
	})
	t.Run("Save", func(t *testing.T) {
		repo, cleanup := f()(t)
		defer cleanup()

		testUserRepositorySave(t, repo)
	})
	t.Run("FindByID", func(t *testing.T) {
		repo, cleanup := f()(t)
		defer cleanup()

		testUserRepositoryFindByID(t, repo)
	})
	t.Run("ExistsByID", func(t *testing.T) {
		repo, cleanup := f()(t)
		defer cleanup()

		testUserRepositoryExistsByID(t, repo)
	})
	t.Run("FindAll", func(t *testing.T) {
		repo, cleanup := f()(t)
		defer cleanup()

		testUserRepositoryFindAll(t, repo)
	})
	t.Run("Count", func(t *testing.T) {
		repo, cleanup := f()(t)
		defer cleanup()

		testUserRepositoryCount(t, repo)
	})
	t.Run("DeleteByID", func(t *testing.T) {
		repo, cleanup := f()(t)
		defer cleanup()

		testUserRepositoryDeleteByID(t, repo)
	})
}

func testUserRepositorySave(t *testing.T, repo user.UserRepository) {
	t.Helper()

	ctx := context.Background()
	validate := testValidator()

	t.Run("nil", func(t *testing.T) {
		_, err := repo.Save(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("empty", func(t *testing.T) {
		entity := &user.User{}

		_, err := repo.Save(ctx, entity)
		assert.Error(t, err)

		inS := schema.UserToSchema(entity)
		validateErr := validate.Struct(inS)

		assert.Contains(t, err.Error(), validateErr.Error())
	})

	t.Run("simple", func(t *testing.T) {
		entity := &user.User{
			ID:                 "0",
			Email:              "user_0@test.com",
			PasswordHash:       "password_0",
			NormalizedUsername: "username_0",
			Username:           "username_0",
		}

		_, err := repo.Save(ctx, entity)
		assert.NoError(t, err)
	})
}

func testUserRepositoryFindByID(t *testing.T, repo user.UserRepository) {
	t.Helper()

	ctx := context.Background()

	users := createUsers(t, repo, 1)

	t.Run("non existent", func(t *testing.T) {
		_, err := repo.FindByID(ctx, "non_existent_id")
		require.Error(t, err)

		assert.True(t, errors.Is(err, database.ErrNotFound))
	})

	t.Run("ok", func(t *testing.T) {
		entity, err := repo.FindByID(ctx, users[0].ID)
		require.NoError(t, err)

		assert.NotNil(t, entity)
		assertUserEqual(t, users[0], entity)
	})
}

func testUserRepositoryExistsByID(t *testing.T, repo user.UserRepository) {
	t.Helper()

	ctx := context.Background()

	users := createUsers(t, repo, 1)

	t.Run("non existent", func(t *testing.T) {
		exists, err := repo.ExistsByID(ctx, "non_existent_id")
		require.NoError(t, err)

		assert.False(t, exists)
	})

	t.Run("ok", func(t *testing.T) {
		exists, err := repo.ExistsByID(ctx, users[0].ID)
		require.NoError(t, err)

		assert.True(t, exists)
	})
}

func testUserRepositoryFindAll(t *testing.T, repo user.UserRepository) {
	t.Helper()

	ctx := context.Background()

	t.Run("empty", func(t *testing.T) {
		results, nextCursor, err := repo.FindAll(ctx, "", 0)
		assert.NoError(t, err)

		assert.Equal(t, 0, len(results))
		assert.Equal(t, "", nextCursor)
	})

	createCount := 7

	users := createUsers(t, repo, createCount)

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
		assert.Equal(t, users[limit-1].ID, nextCursor)

		// next page
		results, nextCursor, err = repo.FindAll(ctx, nextCursor, limit)
		assert.NoError(t, err)

		assert.Equal(t, createCount-limit, len(results))
		assert.Equal(t, "", nextCursor)
	})
}

func testUserRepositoryCount(t *testing.T, repo user.UserRepository) {
	t.Helper()

	ctx := context.Background()

	count, err := repo.Count(ctx)
	require.NoError(t, err)

	assert.Equal(t, 0, count)

	createCount := 3

	createUsers(t, repo, createCount)

	count, err = repo.Count(ctx)
	require.NoError(t, err)

	assert.Equal(t, createCount, count)
}

func testUserRepositoryDeleteByID(t *testing.T, repo user.UserRepository) {
	t.Helper()

	ctx := context.Background()

	users := createUsers(t, repo, 1)

	exists, err := repo.ExistsByID(ctx, users[0].ID)
	require.NoError(t, err)

	assert.True(t, exists)

	t.Run("non existent", func(t *testing.T) {
		err := repo.DeleteByID(ctx, "non_existent_id")
		require.NoError(t, err)

		exists, err = repo.ExistsByID(ctx, users[0].ID)
		require.NoError(t, err)

		assert.True(t, exists)
	})

	t.Run("ok", func(t *testing.T) {
		err := repo.DeleteByID(ctx, users[0].ID)
		require.NoError(t, err)

		exists, err = repo.ExistsByID(ctx, users[0].ID)
		require.NoError(t, err)

		assert.False(t, exists)
	})
}
