package di

import (
	"go.uber.org/fx"
)

var Module = fx.Options(
	repositoresfx,
	usecasesfx,
)
