package tofu

import (
	"strings"

	"github.com/google/uuid"
)

type UUID struct {
}

func NewUUIDGenerator() *UUID {
	u := &UUID{}
	return u
}

func (u *UUID) Generate() string {
	return strings.Replace(uuid.New().String(), "-", "", -1)
}
