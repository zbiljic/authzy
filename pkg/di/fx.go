package di

import (
	"go.uber.org/fx"

	account "github.com/zbiljic/authzy/pkg/domain/account/di"
	refreshtoken "github.com/zbiljic/authzy/pkg/domain/refreshtoken/di"
	user "github.com/zbiljic/authzy/pkg/domain/user/di"
)

var Module = fx.Options(
	configfx,
	validatorfx,
	hasherfx,
	debugfx,
	serverfx,
	databasefx,
	account.Module,
	refreshtoken.Module,
	user.Module,
	jwtfx,
	apifx,
)
