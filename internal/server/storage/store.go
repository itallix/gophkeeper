package storage

// SecretStore defines interface for storing and retrieving secrets.
type SecretStore interface {
	Store(path string) error
	Retrieve(path string) error
	Delete(path string) error
}
