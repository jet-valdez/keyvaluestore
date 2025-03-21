package storage

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestFileTransactionLogger_WriteAndReadEvents(t *testing.T) {
	// 1. Create a temporary file for the transaction log
	tmpFile, err := os.CreateTemp("", "transaction_*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFileName := tmpFile.Name()
	// We only needed the name; close it right away
	tmpFile.Close()
	defer os.Remove(tmpFileName)

	// 2. Create a FileTransactionLogger with this temp file
	logger, err := newFileTransactionLogger(tmpFileName)
	if err != nil {
		t.Fatalf("failed to create newFileTransactionLogger: %v", err)
	}

	fileLogger, ok := logger.(*FileTransactionLogger)
	if !ok {
		t.Fatalf("logger is not *FileTransactionLogger")
	}

	// 3. Start the logger’s internal goroutine
	fileLogger.Run()

	// 4. Write some events
	fileLogger.WritePut("alpha", "1")
	fileLogger.WritePut("beta", "2")
	fileLogger.WriteDelete("alpha")

	// 5. Close the `events` channel to signal we’re done writing
	close(fileLogger.events)

	// 6. Watch for any write errors
	for writeErr := range fileLogger.errors {
		if writeErr != nil {
			t.Fatalf("Got an error from the transaction logger: %v", writeErr)
		}
	}

	// 7. Close the file so we can read it from the beginning
	fileLogger.file.Close()

	// 8. Re-open the file for verification
	f, err := os.Open(tmpFileName)
	if err != nil {
		t.Fatalf("failed to re-open file %s: %v", tmpFileName, err)
	}
	defer f.Close()

	// 9. Read the lines in the file
	scanner := bufio.NewScanner(f)
	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Expect exactly 3 lines in the log
	if len(lines) != 3 {
		t.Fatalf("Expected 3 logged events, got %d lines", len(lines))
	}

	// 10. Parse the lines & check their contents
	//    Format: "<sequence>\t<eventType>\t<key>\t<value>"
	//    e.g. "1    2   alpha   1"
	// Sequence # should increment, eventType matches, key/value match
	// For reference: EventDelete=1, EventPut=2 (based on your iota definition).
	var parsedEvents []Event

	for _, line := range lines {
		var e Event
		_, err := fmt.Sscanf(line, "%d\t%d\t%s\t%s", &e.Sequence, &e.EventType, &e.Key, &e.Value)
		if err != nil {
			t.Fatalf("Failed to parse log line '%s': %v", line, err)
		}
		parsedEvents = append(parsedEvents, e)
	}

	// Check we have the correct data in order
	//  1) EventPut alpha=1
	//  2) EventPut beta=2
	//  3) EventDelete alpha
	expected := []Event{
		{Sequence: 1, EventType: EventPut, Key: "alpha", Value: "1"},
		{Sequence: 2, EventType: EventPut, Key: "beta", Value: "2"},
		{Sequence: 3, EventType: EventDelete, Key: "alpha", Value: ""},
	}

	if !reflect.DeepEqual(parsedEvents, expected) {
		t.Errorf("Unexpected event log content.\nGot:      %#v\nExpected: %#v", parsedEvents, expected)
	}
}

func TestInitializeTransactionLogger_ReplaysIntoDB(t *testing.T) {
	// 1. Create a temp file with some pre-written events
	tmpFile, err := os.CreateTemp("", "transaction_*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFileName := tmpFile.Name()
	defer os.Remove(tmpFileName)

	// 2. Manually write a few events to simulate a prior run
	//    Let’s put "foo=bar", "bob=alice", then delete "foo"
	lines := []string{
		"1\t2\tfoo\tbar",   // Sequence=1, EventPut
		"2\t2\tbob\talice", // Sequence=2, EventPut
		"3\t1\tfoo\t",      // Sequence=3, EventDelete
	}
	content := strings.Join(lines, "\n") + "\n"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed writing to temp file: %v", err)
	}
	tmpFile.Close()

	// 3. Create a new in-memory DB (from your code snippet)
	db, err := NewInMemoryDB()
	if err != nil {
		t.Fatalf("Failed to create inMemoryDB: %v", err)
	}

	// 4. Now call InitializeTransactionLogger, which should:
	//    - open the file
	//    - read & replay the 3 events
	//    - call Run() so it can accept new events
	logger, err := InitializeTransactionLogger(db)
	if err != nil {
		t.Fatalf("InitializeTransactionLogger returned an error: %v", err)
	}

	// 5. Check that the DB is now in the correct state:
	//    "foo" was added then deleted, so it should not exist
	//    "bob" -> "alice" should exist
	if _, err := db.Get("foo"); err == nil {
		t.Error("Expected 'foo' to be deleted, but Get('foo') did not return an error")
	}

	val, err := db.Get("bob")
	if err != nil {
		t.Errorf("Get('bob') returned unexpected error: %v", err)
	} else if *val != "alice" {
		t.Errorf("Expected 'alice' for 'bob', got '%s'", *val)
	}

	// 6. Now that logger is running, let’s do additional writes
	fileLogger, ok := logger.(*FileTransactionLogger)
	if !ok {
		t.Fatalf("logger is not a *FileTransactionLogger")
	}

	fileLogger.WritePut("charlie", "123")
	fileLogger.WriteDelete("bob")

	// Close the events channel
	close(fileLogger.events)

	// Check for errors
	for writeErr := range fileLogger.errors {
		if writeErr != nil {
			t.Fatalf("Got an error from the transaction logger: %v", writeErr)
		}
	}

	// 7. DB should have charlie=123, and bob was just deleted
	vCharlie, err := db.Get("charlie")
	if err != nil {
		t.Errorf("Get('charlie') returned unexpected error: %v", err)
	} else if *vCharlie != "123" {
		t.Errorf("Expected '123' for 'charlie', got '%s'", *vCharlie)
	}

	if _, err := db.Get("bob"); err == nil {
		t.Error("Expected 'bob' to be deleted, but found it in DB")
	}
}
