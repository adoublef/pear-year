package text

import "fmt"

// A Name is the given name of a person.
// It can be formatted as a string with at most, 30 characters.
type Name string

func (n *Name) UnmarshalText(p []byte) (err error) {
	*n, err = ParseName(string(p))
	return
}

func (n *Name) Scan(v any) (err error) {
	switch v := v.(type) {
	case string:
		*n, err = ParseName(v)
	case []byte:
		*n, err = ParseName(string(v))
	default:
		return fmt.Errorf("converting type %T to a text.Name", v)
	}
	return err
}

func ParseName(s string) (Name, error) {
	if n := len(s); n > 30 {
		return "", ErrBadLength
	}
	// trim start and end, trim spaces greater than one character
	return Name(s), nil
}
