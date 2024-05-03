package user

import (
	"errors"

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
	// Bio  string // can be any length
}

func (u User) Age() uint8 {
	now := date.Now()
	age := now.Year - u.DOB.Year
	if now.Month < u.DOB.Month || (now.Month == u.DOB.Month && now.Day < u.DOB.Day) {
		age--
	}
	return uint8(age)
}
