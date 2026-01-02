package validator

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	"github.com/shandysiswandi/gobite/internal/pkg/strcase"
)

var (
	// Based on NIST 800-63B Guidelines
	rePassword = regexp.MustCompile(`^.{8,72}$`)
)

// ErrTranslatorNotFound indicates the requested translator is unavailable.
var ErrTranslatorNotFound = errors.New("translator not found")

// V10Validator implements Validator using go-playground/validator v10.
type V10Validator struct {
	validate   *validator.Validate
	translator ut.Translator
}

// V10ValidationError is a field-to-message map returned when validation fails.
//
// Keys are field names in snake_case to match typical JSON conventions.
type V10ValidationError map[string]string

// Error implements the error interface.
func (vs V10ValidationError) Error() string {
	if len(vs) == 0 {
		return "validation error"
	}

	b, err := json.Marshal(vs)
	if err != nil {
		return fmt.Sprintf("validation error (failed to marshal: %v)", err)
	}
	return string(b)
}

// Values returns the field error map.
func (vs V10ValidationError) Values() map[string]string {
	return vs
}

// NewV10Validator constructs a V10Validator with English translations and custom rules.
func NewV10Validator() (*V10Validator, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	enLang := en.New()
	uni := ut.New(enLang, enLang)
	enTrans, ok := uni.GetTranslator("en")
	if !ok {
		return nil, ErrTranslatorNotFound
	}

	if err := enTranslations.RegisterDefaultTranslations(validate, enTrans); err != nil {
		return nil, err
	}

	v10CustomValidation(validate, enTrans)

	return &V10Validator{
		validate:   validate,
		translator: enTrans,
	}, nil
}

// Validate validates a struct and returns a V10ValidationError on failure.
func (v *V10Validator) Validate(data any) error {
	if err := v.validate.Struct(data); err != nil {
		var validateErrs validator.ValidationErrors
		if !errors.As(err, &validateErrs) {
			return err
		}

		errV10 := make(V10ValidationError)
		for _, fe := range validateErrs {
			errV10[strcase.ToLowerSnake(fe.Field())] = fe.Translate(v.translator)
		}

		return errV10
	}

	return nil
}

//nolint:errcheck,gosec,forcetypeassert // make linter silent
func v10CustomValidation(validate *validator.Validate, enTrans ut.Translator) {
	validate.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		p, ok := fl.Field().Interface().(string)
		if !ok {
			return false
		}

		return rePassword.MatchString(p)
	})

	validate.RegisterTranslation("password", enTrans,
		func(ut ut.Translator) error {
			return ut.Add("password", "{0} must be 8-72 characters", false)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T(fe.Tag(), fe.Field())
			return t
		},
	)

	validate.RegisterTranslation("alphaspace", enTrans,
		func(ut ut.Translator) error {
			return ut.Add("alphaspace", "{0} can contain only letters and spaces", false)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, err := ut.T(fe.Tag(), fe.Field())
			if err != nil {
				slog.Warn("warning: error translating", "FieldError", fe, "error", err)
				return fe.(error).Error()
			}

			return t
		},
	)
}
