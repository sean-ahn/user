package crypto

import (
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/scrypt"
)

type ScryptHasher struct {
	salt   []byte
	params scryptParams
}

var _ Hasher = (*ScryptHasher)(nil)

type scryptParams struct {
	n      int
	r      int
	p      int
	keyLen int
}

func NewScryptHasher(salt []byte) *ScryptHasher {
	return &ScryptHasher{
		salt: salt,
		params: scryptParams{
			// See https://github.com/golang/go/issues/22082
			n:      1 << 15,
			r:      8,
			p:      1,
			keyLen: 32,
		},
	}
}

func (h *ScryptHasher) Hash(s []byte) ([]byte, error) {
	dk, err := scrypt.Key(s, h.salt, h.params.n, h.params.r, h.params.p, h.params.keyLen)
	if err != nil {
		return nil, err
	}

	return []byte(fmt.Sprintf(
		"scrypt$%d$%d$%d$%s$%s",
		h.params.n,
		h.params.r,
		h.params.p,
		base64.StdEncoding.EncodeToString(h.salt),
		base64.StdEncoding.EncodeToString(dk),
	)), nil
}
