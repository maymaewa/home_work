package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	ErrInvalidType   = errors.New("invalid type")
	ErrInvalidTag    = errors.New("invalid validation tag")
	ErrInvalidRegexp = errors.New("invalid regular expression")
	ErrValidation    = errors.New("validation failed")
	ErrLen           = errors.New("length validation failed")
	ErrRegexp        = errors.New("regexp validation failed")
	ErrIn            = errors.New("in validation failed")
	ErrMin           = errors.New("min validation failed")
	ErrMax           = errors.New("max validation failed")
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return ""
	}

	var result strings.Builder

	for i, err := range v {
		if i > 0 {
			result.WriteString("; ")
		}

		result.WriteString(err.Field)
		result.WriteString(": ")
		result.WriteString(err.Err.Error())
	}

	return result.String()
}

func Validate(v interface{}) error {
	value := reflect.ValueOf(v)

	if !value.IsValid() || value.Kind() != reflect.Struct {
		return ErrInvalidType
	}

	var validationErrors ValidationErrors

	for i := 0; i < value.NumField(); i++ {
		fieldType := value.Type().Field(i)

		if fieldType.PkgPath != "" {
			continue
		}

		tag, ok := fieldType.Tag.Lookup("validate")
		if !ok || tag == "" {
			continue
		}

		fieldErrors, err := validateField(
			fieldType.Name,
			value.Field(i),
			tag,
		)
		if err != nil {
			return err
		}

		validationErrors = append(validationErrors, fieldErrors...)
	}

	if len(validationErrors) > 0 {
		return validationErrors
	}

	return nil
}

func validateField(
	fieldName string,
	value reflect.Value,
	tag string,
) (ValidationErrors, error) {
	var validationErrors ValidationErrors

	rules := strings.Split(tag, "|")

	for _, rule := range rules {
		if rule == "" {
			return nil, fmt.Errorf("%w: empty validator", ErrInvalidTag)
		}

		if value.Kind() == reflect.Slice {
			for i := 0; i < value.Len(); i++ {
				err := validateValue(fieldName, value.Index(i), rule)
				if err != nil {
					if errors.Is(err, ErrInvalidTag) ||
						errors.Is(err, ErrInvalidRegexp) {
						return nil, err
					}

					validationErrors = append(
						validationErrors,
						ValidationError{
							Field: fieldName,
							Err:   err,
						},
					)
				}
			}

			continue
		}

		err := validateValue(fieldName, value, rule)
		if err != nil {
			if errors.Is(err, ErrInvalidTag) ||
				errors.Is(err, ErrInvalidRegexp) {
				return nil, err
			}

			validationErrors = append(
				validationErrors,
				ValidationError{
					Field: fieldName,
					Err:   err,
				},
			)
		}
	}

	return validationErrors, nil
}

func validateValue(
	_ string,
	value reflect.Value,
	rule string,
) error {
	parts := strings.SplitN(rule, ":", 2)

	if len(parts) != 2 {
		return fmt.Errorf("%w: %s", ErrInvalidTag, rule)
	}

	validator := parts[0]
	argument := parts[1]

	if value.Kind() == reflect.String {
		return validateString(value.String(), validator, argument)
	}

	if value.Kind() == reflect.Int {
		return validateInt(value.Int(), validator, argument)
	}

	return fmt.Errorf(
		"%w: unsupported field type %s",
		ErrInvalidType,
		value.Kind(),
	)
}

func validateString(
	value string,
	validator string,
	argument string,
) error {
	switch validator {
	case "len":
		expectedLen, err := strconv.Atoi(argument)
		if err != nil || expectedLen < 0 {
			return fmt.Errorf(
				"%w: invalid len value %q",
				ErrInvalidTag,
				argument,
			)
		}

		if utf8.RuneCountInString(value) != expectedLen {
			return fmt.Errorf(
				"%w: expected %d characters, got %d",
				ErrLen,
				expectedLen,
				utf8.RuneCountInString(value),
			)
		}

		return nil

	case "regexp":
		re, err := regexp.Compile(argument)
		if err != nil {
			return fmt.Errorf(
				"%w: %w",
				ErrInvalidRegexp,
				err,
			)
		}

		if !re.MatchString(value) {
			return fmt.Errorf(
				"%w: value %q does not match %q",
				ErrRegexp,
				value,
				argument,
			)
		}

		return nil

	case "in":
		values := strings.Split(argument, ",")

		for _, allowed := range values {
			if value == allowed {
				return nil
			}
		}

		return fmt.Errorf(
			"%w: value %q is not allowed",
			ErrIn,
			value,
		)

	default:
		return fmt.Errorf(
			"%w: unknown validator %q for string",
			ErrInvalidTag,
			validator,
		)
	}
}

func validateInt(
	value int64,
	validator string,
	argument string,
) error {
	switch validator {
	case "min":
		limit, err := strconv.ParseInt(argument, 10, 64)
		if err != nil {
			return fmt.Errorf(
				"%w: invalid min value %q",
				ErrInvalidTag,
				argument,
			)
		}

		if value < limit {
			return fmt.Errorf(
				"%w: %d is less than %d",
				ErrMin,
				value,
				limit,
			)
		}

		return nil

	case "max":
		limit, err := strconv.ParseInt(argument, 10, 64)
		if err != nil {
			return fmt.Errorf(
				"%w: invalid max value %q",
				ErrInvalidTag,
				argument,
			)
		}

		if value > limit {
			return fmt.Errorf(
				"%w: %d is greater than %d",
				ErrMax,
				value,
				limit,
			)
		}

		return nil

	case "in":
		values := strings.Split(argument, ",")

		for _, item := range values {
			allowed, err := strconv.ParseInt(item, 10, 64)
			if err != nil {
				return fmt.Errorf(
					"%w: invalid in value %q",
					ErrInvalidTag,
					item,
				)
			}

			if value == allowed {
				return nil
			}
		}

		return fmt.Errorf(
			"%w: value %d is not allowed",
			ErrIn,
			value,
		)

	default:
		return fmt.Errorf(
			"%w: unknown validator %q for int",
			ErrInvalidTag,
			validator,
		)
	}
}
