package secret

type staticSecret struct {
	value []byte
}

func NewStaticValueSecretFromString(value string) (Secret, error) {
	secret := &staticSecret{value: []byte(value)}
	return secret, nil
}

func (s *staticSecret) Get() ([]byte, error) {
	return s.value, nil
}

func (s *staticSecret) Reload() error {
	return nil
}
