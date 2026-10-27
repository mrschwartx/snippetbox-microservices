package validator

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"snippetbox/backend-snippet/internal/apperror"
)

var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

type Validator struct {
	fieldErrors    map[string]string
	nonFieldErrors []string
}

func (v *Validator) Valid() bool {
	return len(v.fieldErrors) == 0 && len(v.nonFieldErrors) == 0
}

func (v *Validator) AddFieldError(key, message string) {
	if v.fieldErrors == nil {
		v.fieldErrors = make(map[string]string)
	}
	if _, exists := v.fieldErrors[key]; !exists {
		v.fieldErrors[key] = message
	}
}

func (v *Validator) AddNonFieldError(message string) {
	v.nonFieldErrors = append(v.nonFieldErrors, message)
}

func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

func (v *Validator) Err() error {
	if v.Valid() {
		return nil
	}

	fields := make(map[string]string, len(v.fieldErrors))
	for key, message := range v.fieldErrors {
		fields[key] = message
	}

	return &apperror.ValidationError{
		FieldErrors:    fields,
		NonFieldErrors: append([]string(nil), v.nonFieldErrors...),
	}
}

func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

func MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

func MinChars(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

func MaxBytes(value string, n int) bool {
	return len(value) <= n
}

func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

func PermittedValue[T comparable](value T, permitted ...T) bool {
	for _, candidate := range permitted {
		if value == candidate {
			return true
		}
	}
	return false
}
