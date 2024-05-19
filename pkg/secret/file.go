package secret

import (
	"os"
)

type fileSecret struct {
	path  string
	value []byte
}

type FileSecretOptions struct {
	Path string
}

func NewFileSecret(options FileSecretOptions) (Secret, error) {
	secret := &fileSecret{path: options.Path}
	if err := secret.Reload(); err != nil {
		return nil, err
	}
	return secret, nil
}

func (s *fileSecret) Get() ([]byte, error) {
	return s.value, nil
}

func (s *fileSecret) Reload() error {
	f, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}

	s.value = f
	return nil
}
