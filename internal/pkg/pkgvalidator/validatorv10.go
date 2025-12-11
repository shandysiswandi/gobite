package pkgvalidator

import (
	"errors"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/id"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	idTranslations "github.com/go-playground/validator/v10/translations/id"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgstrcase"
)

type V10Validator struct {
	validate    *validator.Validate
	translators map[string]ut.Translator
}

type V10ValidationError map[string]string

func (vs V10ValidationError) Error() string {
	return "go-playground/validator/v10/custom-field-error"
}

func (vs V10ValidationError) Values() map[string]string {
	return vs
}

func NewV10Validator() (*V10Validator, error) {
	enLang := en.New()
	idLang := id.New()

	validate := validator.New(validator.WithRequiredStructEnabled())
	uni := ut.New(enLang, enLang, idLang)

	enTrans, _ := uni.GetTranslator("en")
	idTrans, _ := uni.GetTranslator("id")

	if err := enTranslations.RegisterDefaultTranslations(validate, enTrans); err != nil {
		return nil, err
	}
	if err := idTranslations.RegisterDefaultTranslations(validate, idTrans); err != nil {
		return nil, err
	}

	return &V10Validator{
		validate:    validate,
		translators: map[string]ut.Translator{"en": enTrans, "id": idTrans},
	}, nil
}

func (v *V10Validator) Validate(data any) error {
	if err := v.validate.Struct(data); err != nil {
		var validateErrs validator.ValidationErrors
		if !errors.As(err, &validateErrs) {
			return err
		}

		errV10 := make(V10ValidationError)
		for _, fe := range validateErrs {
			errV10[pkgstrcase.ToLowerSnake(fe.Field())] = fe.Translate(v.translators["en"])
		}

		return errV10
	}

	return nil
}
