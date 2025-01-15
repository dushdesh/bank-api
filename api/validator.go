package api

import (
	"bank/util"

	"github.com/go-playground/validator/v10"
)

var validCurrency validator.Func = func(fl validator.FieldLevel) bool {
	if currency, ok := fl.Field().Interface().(string); ok {
		// check for currentcy support
		return util.IsSupportedCurrency(currency)
	}
	return false
}
