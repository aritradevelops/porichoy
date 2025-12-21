package validation

import (
	"slices"
	"strings"

	"github.com/aritradeveops/porichoy/pkg/timex"
	"github.com/go-playground/validator/v10"
)

// password should contain
// 1. at least 8 characters
// 2. at least one uppercase letter
// 3. at least one lowercase letter
// 4. at least one number
// 5. at least one special character
func ValidatePassword(password string) ValidationErrors {
	var errs ValidationErrors
	if len(password) < 8 {
		errs = append(errs, ValidationError{
			Field: "password",
			Param: "min",
			Code:  "min",
			Value: "8",
		})
	}
	if !strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		errs = append(errs, ValidationError{
			Field: "password",
			Param: "uppercase",
			Code:  "uppercase",
			Value: password,
		})
	}
	if !strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz") {
		errs = append(errs, ValidationError{
			Field: "password",
			Param: 1,
			Code:  "lowercase",
			Value: password,
		})
	}
	if !strings.ContainsAny(password, "0123456789") {
		errs = append(errs, ValidationError{
			Field: "password",
			Param: 1,
			Code:  "number",
			Value: password,
		})
	}
	if !strings.ContainsAny(password, "!@#$%^&*()_+"+"-=[]{}|;:,.<>?") {
		errs = append(errs, ValidationError{
			Field: "password",
			Param: 1,
			Code:  "special",
			Value: password,
		})
	}
	return errs
}

// resolvers can be of the following type
// env:<environment_variable_name>
// db:<table_id>
// literal:<literal_value>
// s3:<s3key>
func ValidateResolvers(fl validator.FieldLevel) bool {
	// will come from resolvers package
	availableResolvers := []string{"env", "db", "literal", "s3"}
	val := fl.Field().String()
	parts := strings.Split(val, ":")
	if len(parts) != 2 {
		return false
	}

	source := parts[0]

	return slices.Contains(availableResolvers, source)
}

func ValidateJWTAlgo(fl validator.FieldLevel) bool {
	availableAlgorithms := []string{"HS256", "HS512", "JWK"}
	val := fl.Field().String()
	return slices.Contains(availableAlgorithms, val)
}

func ValidateDuration(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	return timex.IsValidDuration(val)
}
