package config

import (
	"fmt"
	"github.com/go-playground/validator/v10"
)

type CustomValidator struct {
	Validator *validator.Validate
}

type ValidateError struct {
	Message string `json:"message"`
}

func (m *ValidateError) Error() string {
	return m.Message
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.Validator.Struct(i); err != nil {
		return &ValidateError{Message: fmt.Sprintf("Input validation errors : %v", err.Error())}
	}
	return nil
}
