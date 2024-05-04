package user

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.adoublef.dev/sdk/time/date"
	"go.pear-year.io/text"
)

var (
	ErrNotFound = errors.New("user: not found")
)

type User struct {
	ID   uuid.UUID
	Name text.Name
	DOB  date.Date
	Role Role
	// Bio  string // can be any length
}

func (u User) String() string {
	return fmt.Sprintf("User(role=%s, name=%s, age=%d)", u.Role, u.Name, u.Age())
}

func (u User) Age() uint8 {
	now := date.Now()
	age := now.Year - u.DOB.Year
	if now.Month < u.DOB.Month || (now.Month == u.DOB.Month && now.Day < u.DOB.Day) {
		age--
	}
	return uint8(age)
}
