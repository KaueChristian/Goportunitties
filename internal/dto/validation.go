package dto

import (
	"errors"
	"net/url"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func init() {
	engine, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}

	// Report the wire name of a field, not its Go name, so the client can map an
	// error straight onto the input it sent — `json` for bodies, `form` for
	// query strings.
	engine.RegisterTagNameFunc(func(field reflect.StructField) string {
		for _, tag := range []string{"json", "form"} {
			name := strings.SplitN(field.Tag.Get(tag), ",", 2)[0]
			if name != "" && name != "-" {
				return name
			}
		}
		return field.Name
	})

	_ = engine.RegisterValidation("httpurl", validateHTTPURL)
}

// validateHTTPURL accepts only absolute http(s) URLs. Rejecting other schemes
// server-side is what keeps a "javascript:" link from ever reaching the anchor
// the frontend renders for an opening.
func validateHTTPURL(fl validator.FieldLevel) bool {
	raw := strings.TrimSpace(fl.Field().String())
	if raw == "" {
		// Emptiness is `required`'s job to report, not this rule's.
		return true
	}

	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

// FieldErrors converts validator output into a field -> message map. It returns
// nil for any other error, which is how callers tell a validation failure from
// a malformed body.
func FieldErrors(err error) map[string]string {
	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) {
		return nil
	}

	fields := make(map[string]string, len(validationErrs))
	for _, fieldErr := range validationErrs {
		fields[fieldErr.Field()] = messageFor(fieldErr)
	}
	return fields
}

func messageFor(fieldErr validator.FieldError) string {
	switch fieldErr.Tag() {
	case "required":
		return "campo obrigatório"
	case "httpurl":
		return "informe uma URL começando com http:// ou https://"
	case "min":
		return "deve ter ao menos " + fieldErr.Param() + " caracteres"
	case "max":
		return "deve ter no máximo " + fieldErr.Param() + " caracteres"
	case "gte":
		return "deve ser maior ou igual a " + fieldErr.Param()
	case "lte":
		return "deve ser menor ou igual a " + fieldErr.Param()
	case "oneof":
		return "valor inválido; use um de: " + strings.ReplaceAll(fieldErr.Param(), " ", ", ")
	default:
		return "valor inválido"
	}
}
