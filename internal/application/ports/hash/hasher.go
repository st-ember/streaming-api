package hash

type Hasher interface {
	Hash(password string) (string, error)
	Verify(password, encoded string) bool
}
