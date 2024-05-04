package user

import (
	"database/sql/driver"
	"errors"
	"fmt"
)

type Role int

const (
	Guest Role = iota
	Support
	Admin
)

var roles = []string{
	"Guest",
	"Support",
	"Admin",
}

func (r Role) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

func (r *Role) UnmarshalText(p []byte) (err error) {
	*r, err = ParseRole(string(p))
	return err
}

func (r Role) Value() (driver.Value, error) {
	return r.String(), nil
}

func (r *Role) Scan(v any) (err error) {
	switch v := v.(type) {
	case string:
		*r, err = ParseRole(v)
	default:
		return fmt.Errorf("converting type %T to a user.Role", v)
	}
	return err
}

func (r Role) String() string {
	if Guest <= r && r <= Admin {
		return roles[r]
	}
	buf := make([]byte, 20)
	n := fmtInt(buf, uint64(r))
	return "%!Role(" + string(buf[n:]) + ")"
}

func ParseRole(s string) (Role, error) {
	v, _, err := lookup(roles, s)
	if err != nil {
		return -1, err
	}
	return Role(v), nil
}

// time/time.go#L759
func fmtInt(buf []byte, v uint64) int {
	w := len(buf)
	if v == 0 {
		w--
		buf[w] = '0'
	} else {
		for v > 0 {
			w--
			buf[w] = byte(v%10) + '0'
			v /= 10
		}
	}
	return w
}

// time/format.go#L371
func match(s1, s2 string) bool {
	for i := 0; i < len(s1); i++ {
		c1 := s1[i]
		c2 := s2[i]
		if c1 != c2 {
			// Switch to lower-case; 'a'-'A' is known to be a single bit.
			c1 |= 'a' - 'A'
			c2 |= 'a' - 'A'
			if c1 != c2 || c1 < 'a' || c1 > 'z' {
				return false
			}
		}
	}
	return true
}

// time/format.go#L387
func lookup(tab []string, val string) (int, string, error) {
	for i, v := range tab {
		if len(val) >= len(v) && match(val[0:len(v)], v) {
			return i, val[len(v):], nil
		}
	}
	return -1, val, errBad
}

var errBad = errors.New("bad value for field") // placeholder not passed to user
