package hw02unpackstring

import (
	"errors"
	"strings"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(s string) (string, error) {
	runes := []rune(s)
	var builder strings.Builder
	length := len(runes)

	for i := 0; i < length; i++ {
		r := runes[i]

		if unicode.IsDigit(r) {
			return "", ErrInvalidString
		}

		count := 1
		next := i + 1

		if next < length && unicode.IsDigit(runes[next]) {
			count = int(runes[next] - '0')
			i++
		}

		builder.WriteString(strings.Repeat(string(r), count))
	}

	return builder.String(), nil
}
