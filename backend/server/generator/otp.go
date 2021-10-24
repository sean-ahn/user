package generator

import (
	"math/rand"
	"strconv"
)

type OTPGenerator struct {
	Len int
}

var _ Generator = (*OTPGenerator)(nil)

func (g *OTPGenerator) Generate() string {
	var code string
	for i := 0; i < g.Len; i++ {
		code += strconv.Itoa(rand.Intn(10))
	}
	return code
}
