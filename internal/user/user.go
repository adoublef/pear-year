package user

import (
	"errors"

	"github.com/google/uuid"
	"go.pear-year.io/text"
)

var (
	ErrNotFound = errors.New("user: not found")
)

type User struct {
	ID   uuid.UUID
	Name text.Name
	Age  uint8
	Bio  string // can be any length
}
