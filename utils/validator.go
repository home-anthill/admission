package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// GetErrorMessage creates a list with all wrong fields.
func GetErrorMessage(err error) string {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return " unknown validation error"
	}
	var b strings.Builder
	for _, fe := range validationErrors {
		fmt.Fprintf(&b, " %s", strings.ToLower(fe.Field()))
	}
	return b.String()
}
