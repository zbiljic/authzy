package usecases

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/zbiljic/authzy/pkg/database"
	"github.com/zbiljic/authzy/pkg/domain/user"
	"github.com/zbiljic/authzy/pkg/hash"
	"github.com/zbiljic/authzy/pkg/ulid"
)

type userUsecase struct {
	noopUserUsecase

	hasher     hash.Hasher
	repository user.UserRepository
}

func NewUserUsecase(
	hasher hash.Hasher,
	repository user.UserRepository,
) user.UserUsecase {
	uc := &userUsecase{
		hasher:     hasher,
		repository: repository,
	}
	return uc
}

func (uc *userUsecase) CreateUser(ctx context.Context, entity *user.User) (*user.User, error) {
	// normalize
	entity.NormalizedUsername = strings.ToLower(entity.Username)
	entity.Email = strings.ToLower(entity.Email)

	// check if username is already taken
	has, err := uc.repository.ExistsByIdentifier(ctx, entity.NormalizedUsername)
	if err != nil {
		return nil, err
	}

	if has {
		return nil, database.ErrAlreadyExists
	}

	// check if email is already taken
	has, err = uc.repository.ExistsByIdentifier(ctx, entity.Email)
	if err != nil {
		return nil, err
	}

	if has {
		return nil, database.ErrAlreadyExists
	}

	// generate ID
	entity.ID = ulid.ULID().String()

	if len(entity.Password) > 0 {
		// hash password
		hashedPassword, err := uc.hasher.Generate(ctx, []byte(entity.Password))
		if err != nil {
			return nil, fmt.Errorf("password hash generate: %w", err)
		}

		entity.PasswordHash = string(hashedPassword)

		now := time.Now()
		entity.PasswordUpdatedAt = &now
	}

	// before save
	if entity.ValidSince != nil && entity.ValidSince.IsZero() {
		entity.ValidSince = nil
	}

	return uc.repository.Save(ctx, entity)
}

func (uc *userUsecase) UpdateUser(ctx context.Context, entity *user.User) (*user.User, error) {
	exists, err := uc.repository.ExistsByID(ctx, entity.ID)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, database.ErrNotFound
	}

	return uc.repository.Save(ctx, entity)
}

func (uc *userUsecase) UpdatePassword(ctx context.Context, id string, password []byte) (*user.User, error) {
	user, err := uc.repository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if len(password) > 0 {
		// hash password
		hashedPassword, err := uc.hasher.Generate(ctx, password)
		if err != nil {
			return nil, fmt.Errorf("password hash generate: %w", err)
		}

		user.PasswordHash = string(hashedPassword)

		now := time.Now()
		user.PasswordUpdatedAt = &now
	}

	return uc.repository.Save(ctx, user)
}

func (uc *userUsecase) ConfirmUser(ctx context.Context, id string) (*user.User, error) {
	user, err := uc.repository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user.EmailVerified = true

	now := time.Now()
	user.ValidSince = &now

	// clear confirmation
	user.ConfirmationToken = ""
	user.ConfirmationSentAt = nil

	return uc.repository.Save(ctx, user)
}

func (uc *userUsecase) ConfirmRecovery(ctx context.Context, user *user.User) (*user.User, error) {
	user.RecoveryToken = ""
	user.RecoverySentAt = nil

	return uc.repository.Save(ctx, user)
}

func (uc *userUsecase) ConfirmEmailChange(ctx context.Context, user *user.User) (*user.User, error) {
	user.Email = user.EmailChange
	user.EmailChange = ""
	user.EmailChangeToken = ""
	user.EmailChangeSentAt = nil

	return uc.repository.Save(ctx, user)
}

func (uc *userUsecase) FindUserByID(ctx context.Context, id string) (*user.User, error) {
	return uc.repository.FindByID(ctx, id)
}

func (uc *userUsecase) FindUserByEmail(ctx context.Context, email string) (*user.User, error) {
	return uc.repository.FindByIdentifier(ctx, email)
}

func (uc *userUsecase) FindUserByConfirmationToken(ctx context.Context, token string) (*user.User, error) {
	return uc.repository.FindByConfirmationToken(ctx, token)
}

func (uc *userUsecase) FindUserByRecoveryToken(ctx context.Context, token string) (*user.User, error) {
	return uc.repository.FindByRecoveryToken(ctx, token)
}

func (uc *userUsecase) Authenticate(ctx context.Context, identifier string, password []byte) (*user.User, error) {
	normalizedIdentifier := strings.ToLower(identifier)
	entity, err := uc.repository.FindByIdentifier(ctx, normalizedIdentifier)
	if err != nil {
		return nil, err
	}

	err = uc.hasher.Compare(ctx, password, []byte(entity.PasswordHash))
	if err != nil {
		if errors.Is(err, hash.ErrMismatchedHashAndPassword) {
			return nil, fmt.Errorf("password does not match: %s", identifier)
		}

		return nil, fmt.Errorf("password compare: %w", err)
	}

	return entity, nil
}

func (uc *userUsecase) UserSignedIn(ctx context.Context, entity *user.User, ipAddress net.IP) (*user.User, error) {
	user, err := uc.repository.FindByID(ctx, entity.ID)
	if err != nil {
		return nil, err
	}

	if ipAddress != nil {
		user.LastIP = ipAddress.String()
	}

	now := time.Now()
	user.LastLoginAt = &now

	user.LoginsCount++

	return uc.repository.Save(ctx, user)
}

func (uc *userUsecase) UpdateUserMetaData(ctx context.Context, user *user.User, updates map[string]interface{}) (*user.User, error) {
	if user.UserMetaData == nil || len(user.UserMetaData) == 0 {
		user.UserMetaData = updates
	} else {
		for key, value := range updates {
			if value != nil {
				user.UserMetaData[key] = value
			} else {
				delete(user.UserMetaData, key)
			}
		}
	}

	return uc.repository.Save(ctx, user)
}

func (uc *userUsecase) UpdateAppMetaData(ctx context.Context, user *user.User, updates map[string]interface{}) (*user.User, error) {
	if user.AppMetaData == nil || len(user.AppMetaData) == 0 {
		user.AppMetaData = updates
	} else {
		for key, value := range updates {
			if value != nil {
				user.AppMetaData[key] = value
			} else {
				delete(user.AppMetaData, key)
			}
		}
	}

	return uc.repository.Save(ctx, user)
}
