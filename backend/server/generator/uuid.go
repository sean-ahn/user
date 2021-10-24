package generator

import "github.com/google/uuid"

type UUIDGenerator struct {
}

var _ Generator = (*UUIDGenerator)(nil)

func (g *UUIDGenerator) Generate() string {
	return uuid.New().String()
}
