package store

type SecretStore interface {
	Store(path string) error
	Retrieve(path string) error
	Delete(path string) error
}
