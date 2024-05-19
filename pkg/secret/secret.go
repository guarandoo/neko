package secret

type Secret interface {
	Get() ([]byte, error)
	Reload() error
}
