package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

// InitValidator initializes the validator
func InitValidator() {
	validate = validator.New()
	
	// Register custom validators
	validate.RegisterValidation("password", validatePassword)
	validate.RegisterValidation("username", validateUsername)
}

// ValidateStruct validates a struct using the validator
func ValidateStruct(s interface{}) error {
	if validate == nil {
		InitValidator()
	}
	return validate.Struct(s)
}

// validatePassword validates password strength
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	
	// At least 6 characters
	if len(password) < 6 {
		return false
	}
	
	// At least one letter
	hasLetter, _ := regexp.MatchString(`[a-zA-Z]`, password)
	if !hasLetter {
		return false
	}
	
	// At least one number
	hasNumber, _ := regexp.MatchString(`[0-9]`, password)
	if !hasNumber {
		return false
	}
	
	return true
}

// validateUsername validates username format
func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	
	// 3-20 characters
	if len(username) < 3 || len(username) > 20 {
		return false
	}
	
	// Only alphanumeric characters and underscores
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, username)
	return matched
}

// FormatValidationErrors formats validation errors into a readable string
func FormatValidationErrors(err error) string {
	if err == nil {
		return ""
	}
	
	var errors []string
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, fmt.Sprintf("%s: %s", e.Field(), getValidationMessage(e)))
		}
	} else {
		errors = append(errors, err.Error())
	}
	
	return strings.Join(errors, "; ")
}

// getValidationMessage returns a user-friendly validation message
func getValidationMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s characters", e.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", e.Param())
	case "password":
		return "must contain at least 6 characters with letters and numbers"
	case "username":
		return "must be 3-20 characters and contain only letters, numbers, and underscores"
	case "oneof":
		return fmt.Sprintf("must be one of: %s", e.Param())
	default:
		return fmt.Sprintf("is not valid (%s)", e.Tag())
	}
}
