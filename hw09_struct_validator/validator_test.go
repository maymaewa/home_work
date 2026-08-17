package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"testing"
)

type UserRole string

type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}
)

func TestValidateValidUser(t *testing.T) {
	user := User{
		ID:     "123456789012345678901234567890123456",
		Name:   "John",
		Age:    25,
		Email:  "john@example.com",
		Role:   "admin",
		Phones: []string{"12345678901"},
	}

	err := Validate(user)

	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateInvalidUser(t *testing.T) {
	user := User{
		ID:     "short",
		Name:   "John",
		Age:    10,
		Email:  "invalid-email",
		Role:   "guest",
		Phones: []string{"123"},
	}

	err := Validate(user)

	requireValidationErrors(t, err, ValidationErrors{
		{Field: "ID", Err: ErrLen},
		{Field: "Age", Err: ErrMin},
		{Field: "Email", Err: ErrRegexp},
		{Field: "Role", Err: ErrIn},
		{Field: "Phones", Err: ErrLen},
	})
}

func TestValidateAgeMin(t *testing.T) {
	type TestStruct struct {
		Age int `validate:"min:18"`
	}

	err := Validate(TestStruct{
		Age: 17,
	})

	requireValidationErrors(t, err, ValidationErrors{
		{Field: "Age", Err: ErrMin},
	})
}

func TestValidateAgeMax(t *testing.T) {
	type TestStruct struct {
		Age int `validate:"max:50"`
	}

	err := Validate(TestStruct{
		Age: 51,
	})

	requireValidationErrors(t, err, ValidationErrors{
		{Field: "Age", Err: ErrMax},
	})
}

func TestValidateAgeRange(t *testing.T) {
	type TestStruct struct {
		Age int `validate:"min:18|max:50"`
	}

	tests := []struct {
		name  string
		age   int
		valid bool
	}{
		{
			name:  "below minimum",
			age:   17,
			valid: false,
		},
		{
			name:  "minimum",
			age:   18,
			valid: true,
		},
		{
			name:  "middle",
			age:   30,
			valid: true,
		},
		{
			name:  "maximum",
			age:   50,
			valid: true,
		},
		{
			name:  "above maximum",
			age:   51,
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(TestStruct{
				Age: tt.age,
			})

			if tt.valid && err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			if !tt.valid {
				var validationErrors ValidationErrors

				if !errors.As(err, &validationErrors) {
					t.Fatalf(
						"expected ValidationErrors, got %T: %v",
						err,
						err,
					)
				}
			}
		})
	}
}

func TestValidateStringLen(t *testing.T) {
	type TestStruct struct {
		Value string `validate:"len:5"`
	}

	tests := []struct {
		name  string
		value string
		valid bool
	}{
		{
			name:  "valid",
			value: "hello",
			valid: true,
		},
		{
			name:  "too short",
			value: "hi",
			valid: false,
		},
		{
			name:  "too long",
			value: "hello world",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(TestStruct{
				Value: tt.value,
			})

			if tt.valid {
				if err != nil {
					t.Fatalf("expected nil, got %v", err)
				}

				return
			}

			requireValidationErrors(t, err, ValidationErrors{
				{Field: "Value", Err: ErrLen},
			})
		})
	}
}

func TestValidateRegexp(t *testing.T) {
	type TestStruct struct {
		Email string `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
	}

	tests := []struct {
		name  string
		value string
		valid bool
	}{
		{
			name:  "valid email",
			value: "user@example.com",
			valid: true,
		},
		{
			name:  "invalid email",
			value: "user@example",
			valid: false,
		},
		{
			name:  "invalid email without domain",
			value: "user@",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(TestStruct{
				Email: tt.value,
			})

			if tt.valid {
				if err != nil {
					t.Fatalf("expected nil, got %v", err)
				}

				return
			}

			requireValidationErrors(t, err, ValidationErrors{
				{Field: "Email", Err: ErrRegexp},
			})
		})
	}
}

func TestValidateStringIn(t *testing.T) {
	type TestStruct struct {
		Role string `validate:"in:admin,stuff"`
	}

	tests := []struct {
		name  string
		value string
		valid bool
	}{
		{
			name:  "admin",
			value: "admin",
			valid: true,
		},
		{
			name:  "stuff",
			value: "stuff",
			valid: true,
		},
		{
			name:  "unknown role",
			value: "user",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(TestStruct{
				Role: tt.value,
			})

			if tt.valid {
				if err != nil {
					t.Fatalf("expected nil, got %v", err)
				}

				return
			}

			requireValidationErrors(t, err, ValidationErrors{
				{Field: "Role", Err: ErrIn},
			})
		})
	}
}

func TestValidateIntIn(t *testing.T) {
	type TestStruct struct {
		Code int `validate:"in:200,404,500"`
	}

	tests := []struct {
		name  string
		value int
		valid bool
	}{
		{
			name:  "200",
			value: 200,
			valid: true,
		},
		{
			name:  "404",
			value: 404,
			valid: true,
		},
		{
			name:  "500",
			value: 500,
			valid: true,
		},
		{
			name:  "unknown code",
			value: 201,
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(TestStruct{
				Code: tt.value,
			})

			if tt.valid {
				if err != nil {
					t.Fatalf("expected nil, got %v", err)
				}

				return
			}

			requireValidationErrors(t, err, ValidationErrors{
				{Field: "Code", Err: ErrIn},
			})
		})
	}
}

func TestValidateStringSlice(t *testing.T) {
	type TestStruct struct {
		Phones []string `validate:"len:11"`
	}

	tests := []struct {
		name  string
		value []string
		valid bool
	}{
		{
			name:  "valid",
			value: []string{"12345678901", "09876543210"},
			valid: true,
		},
		{
			name:  "invalid one element",
			value: []string{"12345678901", "123"},
			valid: false,
		},
		{
			name:  "invalid all elements",
			value: []string{"123", "456"},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(TestStruct{
				Phones: tt.value,
			})

			if tt.valid {
				if err != nil {
					t.Fatalf("expected nil, got %v", err)
				}

				return
			}

			var validationErrors ValidationErrors

			if !errors.As(err, &validationErrors) {
				t.Fatalf(
					"expected ValidationErrors, got %T: %v",
					err,
					err,
				)
			}

			if len(validationErrors) == 0 {
				t.Fatal("expected validation errors, got empty slice")
			}
		})
	}
}

func TestValidateIgnoresFieldsWithoutTag(t *testing.T) {
	type TestStruct struct {
		Valid   string `validate:"len:5"`
		Ignored string
	}

	err := Validate(TestStruct{
		Valid:   "hello",
		Ignored: "this value can have any length",
	})

	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateIgnoresUnexportedFields(t *testing.T) {
	user := User{
		ID:    "123456789012345678901234567890123456",
		Age:   25,
		Email: "john@example.com",
		Role:  "admin",
		meta:  json.RawMessage(`invalid`),
	}

	err := Validate(user)

	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateStructWithoutTags(t *testing.T) {
	token := Token{
		Header:    []byte("header"),
		Payload:   []byte("payload"),
		Signature: []byte("signature"),
	}

	err := Validate(token)

	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateInvalidType(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
	}{
		{
			name:  "string",
			value: "string",
		},
		{
			name:  "integer",
			value: 123,
		},
		{
			name:  "slice",
			value: []int{1, 2, 3},
		},
		{
			name:  "nil",
			value: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.value)

			if !errors.Is(err, ErrInvalidType) {
				t.Fatalf(
					"expected ErrInvalidType, got %v",
					err,
				)
			}
		})
	}
}

func TestValidateInvalidRegexp(t *testing.T) {
	type TestStruct struct {
		Value string `validate:"regexp:["`
	}

	err := Validate(TestStruct{
		Value: "hello",
	})

	if !errors.Is(err, ErrInvalidRegexp) {
		t.Fatalf(
			"expected ErrInvalidRegexp, got %v",
			err,
		)
	}
}

func TestValidateInvalidTag(t *testing.T) {
	t.Run("unknown validator", func(t *testing.T) {
		type TestStruct struct {
			Value int `validate:"unknown:10"`
		}

		err := Validate(TestStruct{
			Value: 10,
		})

		if !errors.Is(err, ErrInvalidTag) {
			t.Fatalf("expected ErrInvalidTag, got %v", err)
		}
	})

	t.Run("missing argument", func(t *testing.T) {
		type TestStruct struct {
			Value int `validate:"min"`
		}

		err := Validate(TestStruct{
			Value: 10,
		})

		if !errors.Is(err, ErrInvalidTag) {
			t.Fatalf("expected ErrInvalidTag, got %v", err)
		}
	})

	t.Run("invalid min", func(t *testing.T) {
		type TestStruct struct {
			Value int `validate:"min:abc"`
		}

		err := Validate(TestStruct{
			Value: 10,
		})

		if !errors.Is(err, ErrInvalidTag) {
			t.Fatalf("expected ErrInvalidTag, got %v", err)
		}
	})

	t.Run("invalid max", func(t *testing.T) {
		type TestStruct struct {
			Value int `validate:"max:abc"`
		}

		err := Validate(TestStruct{
			Value: 10,
		})

		if !errors.Is(err, ErrInvalidTag) {
			t.Fatalf("expected ErrInvalidTag, got %v", err)
		}
	})

	t.Run("invalid len", func(t *testing.T) {
		type TestStruct struct {
			Value string `validate:"len:abc"`
		}

		err := Validate(TestStruct{
			Value: "hello",
		})

		if !errors.Is(err, ErrInvalidTag) {
			t.Fatalf("expected ErrInvalidTag, got %v", err)
		}
	})
}

func TestValidationErrorsError(t *testing.T) {
	err := ValidationErrors{
		{
			Field: "Name",
			Err:   ErrLen,
		},
		{
			Field: "Age",
			Err:   ErrMin,
		},
	}

	expected := "Name: length validation failed; Age: min validation failed"

	if err.Error() != expected {
		t.Fatalf(
			"expected %q, got %q",
			expected,
			err.Error(),
		)
	}
}

func requireValidationErrors(
	t *testing.T,
	err error,
	expected ValidationErrors,
) {
	t.Helper()

	var actual ValidationErrors

	if !errors.As(err, &actual) {
		t.Fatalf(
			"expected ValidationErrors, got %T: %v",
			err,
			err,
		)
	}

	if len(actual) != len(expected) {
		t.Fatalf(
			"expected %d validation errors, got %d",
			len(expected),
			len(actual),
		)
	}

	for i := range expected {
		if actual[i].Field != expected[i].Field {
			t.Errorf(
				"error %d: expected field %q, got %q",
				i,
				expected[i].Field,
				actual[i].Field,
			)
		}

		if !errors.Is(actual[i].Err, expected[i].Err) {
			t.Errorf(
				"error %d: expected %v, got %v",
				i,
				expected[i].Err,
				actual[i].Err,
			)
		}
	}
}