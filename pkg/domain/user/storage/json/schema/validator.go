package schema

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

const (
	usernameRegexString = "^[a-zA-Z0-9][a-zA-Z0-9_-]{2,31}$"
)

var (
	usernameRegex = regexp.MustCompile(usernameRegexString)

	validators = map[string]validator.Func{
		"username": isUsername,
	}
)

func RegisterValidators(v *validator.Validate) {
	for k, val := range validators {
		_ = v.RegisterValidation(k, val)
	}
}

func isUsername(fl validator.FieldLevel) bool {
	return usernameRegex.MatchString(fl.Field().String())
}
