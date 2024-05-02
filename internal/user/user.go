package user

import (
	"github.com/google/uuid"
	"go.pear-year.io/text"
)

type User struct {
	ID   uuid.UUID
	Name text.Name
	Age  uint8
	Bio  string // can be any length
}
