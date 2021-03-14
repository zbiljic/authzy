package usecases

import (
	"context"
	"net"

	"github.com/zbiljic/authzy/pkg/domain/user"
)

// Compile-time proof of interface implementation.
var _ user.UserUsecase = (*noopUserUsecase)(nil)

// noopUserUsecase can be embedded to have forward compatible implementations.
type noopUserUsecase struct{}

func (*noopUserUsecase) CreateUser(ctx context.Context, user *user.User) (*user.User, error) {
	panic("CreateUser not implemented")
}

func (*noopUserUsecase) UpdateUser(ctx context.Context, user *user.User) (*user.User, error) {
	panic("UpdateUser not implemented")
}

func (*noopUserUsecase) UpdatePassword(ctx context.Context, id string, password []byte) (*user.User, error) {
	panic("UpdatePassword not implemented")
}

func (*noopUserUsecase) ConfirmUser(ctx context.Context, id string) (*user.User, error) {
	panic("ConfirmUser not implemented")
}

func (*noopUserUsecase) ConfirmRecovery(ctx context.Context, user *user.User) (*user.User, error) {
	panic("ConfirmRecovery not implemented")
}

func (*noopUserUsecase) ConfirmEmailChange(ctx context.Context, user *user.User) (*user.User, error) {
	panic("ConfirmEmailChange not implemented")
}

func (*noopUserUsecase) FindUserByID(ctx context.Context, id string) (*user.User, error) {
	panic("FindUserByID not implemented")
}

func (*noopUserUsecase) FindUserByEmail(ctx context.Context, email string) (*user.User, error) {
	panic("FindUserByEmail not implemented")
}

func (*noopUserUsecase) FindUserByConfirmationToken(ctx context.Context, token string) (*user.User, error) {
	panic("FindUserByConfirmationToken not implemented")
}

func (*noopUserUsecase) FindUserByRecoveryToken(ctx context.Context, token string) (*user.User, error) {
	panic("FindUserByRecoveryToken not implemented")
}

func (*noopUserUsecase) Authenticate(ctx context.Context, identifier string, password []byte) (*user.User, error) {
	panic("Authenticate not implemented")
}

func (*noopUserUsecase) UserSignedIn(ctx context.Context, user *user.User, ipAddress net.IP) (*user.User, error) {
	panic("UserSignedIn not implemented")
}

func (*noopUserUsecase) UpdateUserMetaData(ctx context.Context, user *user.User, updates map[string]interface{}) (*user.User, error) {
	panic("UpdateUserMetaData not implemented")
}

func (*noopUserUsecase) UpdateAppMetaData(ctx context.Context, user *user.User, updates map[string]interface{}) (*user.User, error) {
	panic("UpdateAppMetaData not implemented")
}
