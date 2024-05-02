package text

import (
	"errors"
	"unicode"
)

var (
	ErrBadLength       = errors.New("text: bad length")
	ErrFirstLetter     = errors.New("text: first letter is not a letter")
	ErrNotAlphaNumeric = errors.New("text: not alphanumeric")
)

func isAlphanumeric(s string) bool {
	for _, char := range s {
		if !unicode.IsLower(char) && !unicode.IsDigit(char) {
			return false
		}
	}
	return true
}
