package storage

import (
	"errors"
	"sync"
)

// sentinel error
var ErrorNoSuchKey = errors.New("no such key")

type inMemoryDB struct {
	store map[string]string
	lck   sync.RWMutex
}

func NewInMemoryDB() (DB, error) {
	return &inMemoryDB{store: make(map[string]string, 0)}, nil
}

// GetAll returns a copy of the underlying store to avoid race conditions.
func (db *inMemoryDB) GetAll() (map[string]string, error) {
	db.lck.RLock()
	defer db.lck.RUnlock()

	copyStore := make(map[string]string, len(db.store))
	for k, v := range db.store {
		copyStore[k] = v
	}
	return copyStore, nil
}

// Get returns the value for a given key if it exists; otherwise, it returns an error.
func (db *inMemoryDB) Get(key string) (*string, error) {
	db.lck.RLock()
	defer db.lck.RUnlock()

	value, ok := db.store[key]
	if !ok {
		return nil, ErrorNoSuchKey
	}
	// Return a pointer to a copy of the value.
	val := value
	return &val, nil
}

// Set stores the key/value pair and returns a pointer to the value.
func (db *inMemoryDB) Upsert(key string, value string) error {
	db.lck.Lock()
	defer db.lck.Unlock()

	db.store[key] = value
	return nil
}

// Delete removes a key from the store if it exists, otherwise it returns an error.
func (db *inMemoryDB) Delete(key string) error {
	db.lck.Lock()
	defer db.lck.Unlock()

	if _, exists := db.store[key]; !exists {
		return errors.New("key not found")
	}
	delete(db.store, key)
	return nil
}
