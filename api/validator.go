package api

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var ratePattern = regexp.MustCompile(`^(?:[0-9]|[1-9][0-9])(?:\.[0-9]{1,})?$`)

var isDiscountRateValid validator.Func = func(FieldLevel validator.FieldLevel) bool {
	if rate, ok := FieldLevel.Field().Interface().(string); ok {
		return ratePattern.MatchString(rate)
	}
	return false
}
