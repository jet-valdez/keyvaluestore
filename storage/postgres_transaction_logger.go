package storage

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/lib/pq"
)

type PostgresTransactionLogger struct {
	events chan<- Event
	errors <-chan error
	db     *sql.DB
	wg     *sync.WaitGroup
}

func NewPostgresTransactionLogger(param PostgresConfig) (TransactionLogger, error) {
	connStr := fmt.Sprintf("host=%s dbname=%s user=%s password=%s sslmode=disable",
		param.host, param.dbName, param.user, param.password)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create db value: %w", err)
	}

	err = db.Ping() // Test the databases connection
	if err != nil {
		return nil, fmt.Errorf("failed to opendb connection: %w", err)
	}

	tl := &PostgresTransactionLogger{db: db, wg: &sync.WaitGroup{}}

	exists, err := tl.verifyTableExists()
	if err != nil {
		return nil, fmt.Errorf("failed to verify table exists: %w", err)
	}
	if !exists {
		if err = tl.createTable(); err != nil {
			return nil, fmt.Errorf("failed to create table: %w", err)
		}
	}

	return tl, nil
}

func (l *PostgresTransactionLogger) WritePut(key, value string) {
	l.events <- Event{EventType: EventPut, Key: key, Value: value}
}

func (l *PostgresTransactionLogger) WriteDelete(key string) {
	l.events <- Event{EventType: EventDelete, Key: key}
}

func (l *PostgresTransactionLogger) Err() <-chan error {
	return l.errors
}

func (l *PostgresTransactionLogger) Run() {
	events := make(chan Event, 16) // Make an events channel
	l.events = events

	errors := make(chan error, 1) // Make an errors channel
	l.errors = errors

	go func() { // The INSERT query
		query := `INSERT INTO transactions
			(event_type, key, value)
			VALUES ($1, $2, $3)`

		for e := range events { // Retrieve the next Event
			_, err := l.db.Exec( // Execute the INSERT query
				query,
				e.EventType, e.Key, e.Value)

			if err != nil {
				errors <- err
			}

			l.wg.Done()
		}
	}()
}

func (l *PostgresTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	outEvent := make(chan Event)    // An unbuffered events channel
	outError := make(chan error, 1) // A buffered errors channel

	query := "SELECT sequence, event_type, key, value FROM transactions"

	go func() {
		defer close(outEvent) // Close the channels when the
		defer close(outError) // goroutine ends

		rows, err := l.db.Query(query) // Run query; get result set
		if err != nil {
			outError <- fmt.Errorf("sql query error: %w", err)
			return
		}

		defer rows.Close() // This is important!

		var e Event // Create an empty Event

		for rows.Next() { // Iterate over the rows

			err = rows.Scan( // Read the values from the
				&e.Sequence, &e.EventType, // row into the Event.
				&e.Key, &e.Value)

			if err != nil {
				outError <- err
				return
			}

			outEvent <- e // Send e to the channel
		}

		err = rows.Err()
		if err != nil {
			outError <- fmt.Errorf("transaction log read failure: %w", err)
		}
	}()

	return outEvent, outError
}

func (l *PostgresTransactionLogger) Wait() {
	l.wg.Wait()
}

func (l *PostgresTransactionLogger) verifyTableExists() (bool, error) {
	const table = "transactions"

	var result string

	rows, err := l.db.Query(fmt.Sprintf("SELECT to_regclass('public.%s');", table))
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() && result != table {
		rows.Scan(&result)
	}

	return result == table, rows.Err()
}

func (l *PostgresTransactionLogger) createTable() error {
	var err error

	createQuery := `CREATE TABLE transactions (
		sequence      BIGSERIAL PRIMARY KEY,
		event_type    SMALLINT,
		key 		  TEXT,
		value         TEXT
	  );`

	_, err = l.db.Exec(createQuery)
	if err != nil {
		return err
	}

	return nil
}
