package storage

import (
	"reflect"
	"testing"
)

// TestInMemoryDB_GetAll tests that GetAll() returns a copy of the store,
// and that modifying the returned map does not affect the original data.
func TestInMemoryDB_GetAll(t *testing.T) {
	db, err := NewInMemoryDB()
	if err != nil {
		t.Fatalf("Failed to create inMemoryDB: %v", err)
	}

	// Insert some key-value pairs.
	err = db.Upsert("key1", "value1")
	if err != nil {
		t.Fatalf("Upsert returned error: %v", err)
	}
	err = db.Upsert("key2", "value2")
	if err != nil {
		t.Fatalf("Upsert returned error: %v", err)
	}

	// Retrieve all data
	allData, err := db.GetAll()
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}

	// Check we got back the entries we expect.
	expected := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	if !reflect.DeepEqual(allData, expected) {
		t.Errorf("GetAll mismatch.\nGot:      %#v\nExpected: %#v", allData, expected)
	}

	// Ensure that modifying the returned map does not affect the underlying store
	allData["key3"] = "should-not-exist"
	stillAllData, err := db.GetAll()
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if _, exists := stillAllData["key3"]; exists {
		t.Errorf("Expected 'key3' not to exist in the real store")
	}
}

// TestInMemoryDB_Get tests retrieving values and the error behavior
// when a key does not exist.
func TestInMemoryDB_Get(t *testing.T) {
	db, err := NewInMemoryDB()
	if err != nil {
		t.Fatalf("Failed to create inMemoryDB: %v", err)
	}

	// Attempt to get a non-existing key
	_, err = db.Get("unknownKey")
	if err == nil {
		t.Error("Expected an error when getting a non-existent key, but got none")
	}
	if err != ErrorNoSuchKey {
		t.Errorf("Expected ErrorNoSuchKey, got: %v", err)
	}

	// Insert a key and then retrieve it
	err = db.Upsert("existingKey", "someValue")
	if err != nil {
		t.Fatalf("Upsert returned error: %v", err)
	}

	value, err := db.Get("existingKey")
	if err != nil {
		t.Fatalf("Get returned error for existing key: %v", err)
	}
	if value == nil || *value != "someValue" {
		t.Errorf("Expected 'someValue', got '%v'", value)
	}
}

// TestInMemoryDB_Upsert tests inserting new keys and updating existing keys.
func TestInMemoryDB_Upsert(t *testing.T) {
	db, err := NewInMemoryDB()
	if err != nil {
		t.Fatalf("Failed to create inMemoryDB: %v", err)
	}

	// Insert a new key
	err = db.Upsert("newKey", "newValue")
	if err != nil {
		t.Fatalf("Upsert returned error: %v", err)
	}

	value, err := db.Get("newKey")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if *value != "newValue" {
		t.Errorf("Expected 'newValue', got '%s'", *value)
	}

	// Update the existing key
	err = db.Upsert("newKey", "updatedValue")
	if err != nil {
		t.Fatalf("Upsert returned error while updating: %v", err)
	}

	value, err = db.Get("newKey")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if *value != "updatedValue" {
		t.Errorf("Expected 'updatedValue', got '%s'", *value)
	}
}

// TestInMemoryDB_Delete tests deleting keys and ensures
// an error is returned for non-existent keys.
func TestInMemoryDB_Delete(t *testing.T) {
	db, err := NewInMemoryDB()
	if err != nil {
		t.Fatalf("Failed to create inMemoryDB: %v", err)
	}

	// Insert and delete a key
	err = db.Upsert("delKey", "delValue")
	if err != nil {
		t.Fatalf("Upsert returned error: %v", err)
	}

	err = db.Delete("delKey")
	if err != nil {
		t.Fatalf("Delete returned error for an existing key: %v", err)
	}

	// Attempt to delete again
	err = db.Delete("delKey")
	if err == nil {
		t.Error("Expected an error when deleting a non-existent key, but got none")
	}
	if err.Error() != "key not found" {
		t.Errorf("Expected 'key not found' error, got '%v'", err)
	}

	// Confirm that 'delKey' is really gone
	_, getErr := db.Get("delKey")
	if getErr == nil {
		t.Error("Expected an error when getting a deleted key, but got none")
	}
	if getErr != ErrorNoSuchKey {
		t.Errorf("Expected ErrorNoSuchKey for deleted key, got '%v'", getErr)
	}
}
