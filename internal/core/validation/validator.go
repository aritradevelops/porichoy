package validation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

func init() {
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})
	validate.RegisterValidation("jwt_algo", ValidateJWTAlgo)
	validate.RegisterValidation("duration", ValidateDuration)
	validate.RegisterValidation("resolver", ValidateResolvers)
}

type ValidationError struct {
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
	Field   string `json:"field,omitempty"`
	Value   any    `json:"value,omitempty"`
	Param   any    `json:"param,omitempty"`
}

type ValidationErrors []ValidationError

func Validate(obj any) ValidationErrors {
	err := validate.Struct(obj)
	if err != nil {
		out := ValidationErrors{}
		for _, e := range err.(validator.ValidationErrors) {
			out = append(out, ValidationError{
				Message: e.Error(),
				Field:   e.Field(),
				Code:    e.Tag(),
				Value:   e.Value(),
				Param:   e.Param(),
			})
		}

		return out
	}
	return nil
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("Field: %s, Code: %s, Value: %v, Param: %v", e.Field, e.Code, e.Value, e.Param)
}

func (es ValidationErrors) Error() string {
	out := []string{}
	for _, e := range es {
		out = append(out, e.Error())
	}
	return strings.Join(out, ", ")
}
