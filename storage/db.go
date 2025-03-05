package storage

type DB interface {
	GetAll() (map[string]string, error)
	Get(key string) (*string, error)
	Upsert(key string, value string) error
	Delete(key string) error
}
