package mdb

import (
	"database/sql"
	"log"
	"time"

	"github.com/mattn/go-sqlite3"
)

type EmailEntry struct {
	Id						int64
	Email				string
	ConfirmedAt				*time.Time
	OptOut					bool
}

func TryCreate(db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE emails (
			id				INTERGER PRIMARY KEY,
			email			TEXT UNIQUE,
			confirmed_at	INTEGER,
			opt_out			INTEGER
		)
	`)
	if err != nil {
		// 'err.(sqlite3.Error)' means the error is being casted to a sqlite3 error
		if sqlError, ok := err.(sqlite3.Error); ok {
			// code 1 == "table already exists"
			if sqlError.Code != 1 {
				log.Fatal(sqlError)
			}
		} else {
			log.Fatal(err)
		}
	}
}

func emailEntryFromRow(row *sql.Rows) (*EmailEntry, error) {
	var id int64
	var email string
	var confirmedAt int64
	var optOut bool

	err := row.Scan(&id, &email, &confirmedAt, &optOut)
	
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// we're storing the time in the database as unix time, which
	// is integers, so we need to convert time (int to time.Time)
	t := time.Unix(confirmedAt, 0)
	return &EmailEntry{Id: id, Email: email, ConfirmedAt: &t, OptOut: optOut}, nil
}