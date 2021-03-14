package test

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zbiljic/authzy/pkg/database"
	"github.com/zbiljic/authzy/pkg/domain/account"
	"github.com/zbiljic/authzy/pkg/domain/account/storage/json/schema"
	"github.com/zbiljic/authzy/pkg/domain/account/storage/json/transformer"
)

const (
	TestPrefix = "test-"
)

func testValidator() *validator.Validate {
	return validator.New()
}

func createAccounts(t *testing.T, repo account.AccountRepository, userID string, count int) []*account.Account {
	t.Helper()

	var result []*account.Account

	ctx := context.Background()

	i := 0
	for i < count {
		i++

		entity := &account.Account{
			UserID:      userID,
			Provider:    account.ProviderTypePassword,
			FederatedID: strconv.FormatInt(int64(i), 10),
		}

		savedEntity, err := repo.Save(ctx, entity)
		assert.NoError(t, err)

		result = append(result, savedEntity)
	}

	return result
}

func assertAccountEqual(t *testing.T, expected, actual *account.Account) {
	t.Helper()

	assert.Equal(t, expected.UserID, actual.UserID)
	assert.Equal(t, expected.Provider, actual.Provider)
	assert.Equal(t, expected.FederatedID, actual.FederatedID)
}

func Run(t *testing.T, f func() func(t *testing.T) (account.AccountRepository, func())) {
	t.Helper()

	t.Run("init", func(t *testing.T) {
		_, cleanup := f()(t)
		defer cleanup()
	})
	t.Run("Save", func(t *testing.T) {
		repo, cleanup := f()(t)
		defer cleanup()

		testAccountRepositorySave(t, repo)
	})
	t.Run("Find", func(t *testing.T) {
		repo, cleanup := f()(t)
		defer cleanup()

		testAccountRepositoryFind(t, repo)
	})
	t.Run("Exists", func(t *testing.T) {
		repo, cleanup := f()(t)
		defer cleanup()

		testAccountRepositoryExists(t, repo)
	})
	t.Run("FindAll", func(t *testing.T) {
		repo, cleanup := f()(t)
		defer cleanup()

		testAccountRepositoryFindAll(t, repo)
	})
	t.Run("Count", func(t *testing.T) {
		repo, cleanup := f()(t)
		defer cleanup()

		testAccountRepositoryCount(t, repo)
	})
	t.Run("Delete", func(t *testing.T) {
		repo, cleanup := f()(t)
		defer cleanup()

		testAccountRepositoryDelete(t, repo)
	})
	t.Run("FindAllForUser", func(t *testing.T) {
		repo, cleanup := f()(t)
		defer cleanup()

		testAccountRepositoryFindAllForUser(t, repo)
	})
}

func testAccountRepositorySave(t *testing.T, repo account.AccountRepository) {
	t.Helper()

	ctx := context.Background()
	validate := testValidator()

	t.Run("nil", func(t *testing.T) {
		_, err := repo.Save(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("empty", func(t *testing.T) {
		entity := &account.Account{}

		_, err := repo.Save(ctx, entity)
		assert.Error(t, err)

		inS := schema.AccountToSchema(entity)
		validateErr := validate.Struct(inS)

		assert.Contains(t, err.Error(), validateErr.Error())
	})

	t.Run("simple", func(t *testing.T) {
		entity := &account.Account{
			UserID:      "0",
			Provider:    account.ProviderTypePassword,
			FederatedID: "0",
		}

		_, err := repo.Save(ctx, entity)
		assert.NoError(t, err)
	})
}

func testAccountRepositoryFind(t *testing.T, repo account.AccountRepository) {
	t.Helper()

	ctx := context.Background()
	userID := "test"

	accounts := createAccounts(t, repo, userID, 1)

	t.Run("non existent", func(t *testing.T) {
		entity := &account.Account{
			UserID:      "nonexistentid",
			Provider:    account.ProviderTypePassword,
			FederatedID: "non_existent_id",
		}

		_, err := repo.Find(ctx, entity)
		require.Error(t, err)

		assert.True(t, errors.Is(err, database.ErrNotFound))
	})

	t.Run("ok", func(t *testing.T) {
		entity, err := repo.Find(ctx, accounts[0])
		require.NoError(t, err)

		assert.NotNil(t, entity)
		assertAccountEqual(t, accounts[0], entity)
	})
}

func testAccountRepositoryExists(t *testing.T, repo account.AccountRepository) {
	t.Helper()

	ctx := context.Background()
	userID := "test"

	accounts := createAccounts(t, repo, userID, 1)

	t.Run("non existent", func(t *testing.T) {
		entity := &account.Account{
			UserID:      "nonexistentid",
			Provider:    account.ProviderTypePassword,
			FederatedID: "non_existent_id",
		}

		exists, err := repo.Exists(ctx, entity)
		require.NoError(t, err)

		assert.False(t, exists)
	})

	t.Run("ok", func(t *testing.T) {
		exists, err := repo.Exists(ctx, accounts[0])
		require.NoError(t, err)

		assert.True(t, exists)
	})
}

func testAccountRepositoryFindAll(t *testing.T, repo account.AccountRepository) {
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

	accounts := createAccounts(t, repo, userID, createCount)

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

		lastAccountSchema := schema.AccountToSchema(accounts[limit-1])
		assert.Equal(t, transformer.MarshalAccountID(lastAccountSchema), nextCursor)

		// next page
		results, nextCursor, err = repo.FindAll(ctx, nextCursor, limit)
		assert.NoError(t, err)

		assert.Equal(t, createCount-limit, len(results))
		assert.Equal(t, "", nextCursor)
	})
}

func testAccountRepositoryCount(t *testing.T, repo account.AccountRepository) {
	t.Helper()

	ctx := context.Background()
	userID := "test"

	count, err := repo.Count(ctx)
	require.NoError(t, err)

	assert.Equal(t, 0, count)

	createCount := 3

	createAccounts(t, repo, userID, createCount)

	count, err = repo.Count(ctx)
	require.NoError(t, err)

	assert.Equal(t, createCount, count)
}

func testAccountRepositoryDelete(t *testing.T, repo account.AccountRepository) {
	t.Helper()

	ctx := context.Background()
	userID := "test"

	accounts := createAccounts(t, repo, userID, 1)

	exists, err := repo.Exists(ctx, accounts[0])
	require.NoError(t, err)

	assert.True(t, exists)

	t.Run("non existent", func(t *testing.T) {
		entity := &account.Account{
			UserID:      "nonexistentid",
			Provider:    account.ProviderTypePassword,
			FederatedID: "non_existent_id",
		}

		err := repo.Delete(ctx, entity)
		require.NoError(t, err)

		exists, err = repo.Exists(ctx, accounts[0])
		require.NoError(t, err)

		assert.True(t, exists)
	})

	t.Run("ok", func(t *testing.T) {
		err := repo.Delete(ctx, accounts[0])
		require.NoError(t, err)

		exists, err = repo.Exists(ctx, accounts[0])
		require.NoError(t, err)

		assert.False(t, exists)
	})
}

func testAccountRepositoryFindAllForUser(t *testing.T, repo account.AccountRepository) {
	t.Helper()

	ctx := context.Background()
	user1ID := "test1"
	user2ID := "test2"

	t.Run("empty", func(t *testing.T) {
		results, err := repo.FindAllForUser(ctx, "")
		assert.Error(t, err)

		assert.Equal(t, 0, len(results))
	})

	createCountUser1 := 2
	createCountUser2 := 1

	var accounts []*account.Account

	accounts = append(accounts, createAccounts(t, repo, user1ID, createCountUser1)...)
	accounts = append(accounts, createAccounts(t, repo, user2ID, createCountUser2)...)

	t.Run("ok", func(t *testing.T) {
		results1, err := repo.FindAllForUser(ctx, user1ID)
		assert.NoError(t, err)
		assert.Equal(t, createCountUser1, len(results1))

		results2, err := repo.FindAllForUser(ctx, user2ID)
		assert.NoError(t, err)
		assert.Equal(t, createCountUser2, len(results2))

		assert.Equal(t, len(accounts), len(results1)+len(results2))
	})
}
