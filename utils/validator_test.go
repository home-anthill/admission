package utils

import (
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
)

type validationFixture struct {
	Name string `validate:"required"`
}

func TestGetErrorMessageListsInvalidFields(t *testing.T) {
	validate := validator.New()
	err := validate.Struct(validationFixture{})

	message := GetErrorMessage(err)

	if message != " name" {
		t.Fatalf("GetErrorMessage() = %q, want %q", message, " name")
	}
}

func TestGetErrorMessageHandlesUnknownError(t *testing.T) {
	message := GetErrorMessage(errors.New("boom"))

	if message != " unknown validation error" {
		t.Fatalf("GetErrorMessage() = %q, want %q", message, " unknown validation error")
	}
}
