package di

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
)

var validatorfx = fx.Provide(
	ProvideValidator,
)

func ProvideValidator() *validator.Validate {
	return validator.New()
}
