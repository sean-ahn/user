package crypto

type Hasher interface {
	Hash(s []byte) ([]byte, error)
}
