package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScryptHasher_Hash(t *testing.T) {
	cases := []struct {
		given    []byte
		expected string
	}{
		{
			given:    []byte("P@ssw0rd"),
			expected: "scrypt$32768$8$1$ciyZXjxl+ByA0jx70OhK/Q==$LEWwE8woAWO5fCb5JjoSZN/0TJP6nHAWa45Jm0jXBM8=",
		},
		{
			given:    []byte("a"),
			expected: "scrypt$32768$8$1$ciyZXjxl+ByA0jx70OhK/Q==$EXiDjdQ9kCanAz6nukxKjgqxBsVWQCe/SvR9I2gJrtU=",
		},
		{
			given:    []byte("5qKH'W~:F\\=;'YBJ{$,<JdH$_Gh'qf54~/B*~ZQN].$5L!\"cYm"),
			expected: "scrypt$32768$8$1$ciyZXjxl+ByA0jx70OhK/Q==$FNTxfTLMMYUjHZ9RhufyjXul9JWUS3jC6sz5+RcjpOE=",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(string(tc.given), func(t *testing.T) {
			hasher := ScryptHasher{
				salt:   []byte{0x72, 0x2c, 0x99, 0x5e, 0x3c, 0x65, 0xf8, 0x1c, 0x80, 0xd2, 0x3c, 0x7b, 0xd0, 0xe8, 0x4a, 0xfd},
				params: scryptParams{n: 1 << 15, r: 8, p: 1, keyLen: 32},
			}

			hash, err := hasher.Hash(tc.given)

			assert.NoError(t, err)
			assert.Equal(t, tc.expected, string(hash))
			assert.Len(t, string(hash), 86)
		})
	}
}
