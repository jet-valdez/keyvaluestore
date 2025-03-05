package storage

type DB interface {
	ReadAll() (map[string]string, error)
	Read(key string) (*string, error)
	Upsert(key string, value string) error
	Delete(key string) error
}
