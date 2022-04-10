package mdb

import (
	"database/sql"
	"log"
	"time"

	"github.com/mattn/go-sqlite3"
)

type EmailEntry struct {
	Id						int64
	Email					string
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

// creates an email data structure from a database row
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

func CreateEmail(db *sql.DB, email string) error {
	// email will replace the '?'
	// '0' is the confirmed_at time, and 0 indicates that the email has not been confirmed
	// opt_out is defaulted to false
	// the id is set automatically
	_, err := db.Exec(`INSERT INTO
		emails(email, confirmed_at, opt_out)
		VALUES(?, 0, false)`, email)
	
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// function to reading an email
func GetEmail(db *sql.DB, email string) (*EmailEntry, error) {
	rows, err := db.Query(`
		SELECT id, email, confirmed_at, opt_out
		FROM emails
		WHERE email = ?`, email)

	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	// email is unique (as defined in the type above) and there should therefore only be one row
	for rows.Next() {
		return emailEntryFromRow(rows)
	}
	// no email found, so return nil, nil
	return nil, nil
}

// UpdateEmail is an 'upsert' function. It will create a new email if
// is already exists. If not, it will do an 'update' operation, setting 
// confirmed_at and op_out. So we never change the id nor email 
func UpdateEmail(db *sql.DB, entry EmailEntry) error {
	t := entry.ConfirmedAt.Unix()

	_, err := db.Exec(`INSERT INTO
		emails(email, confirmed_at, opt_out)
		VALUES(?, ?, ?)
		ON CONFLICT(email) DO UPDATE SET
			confirmed_at=?
			opt_out=?
	`, entry.Email, t, entry.OptOut, t, entry.OptOut)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// Since this is a mailing list, we don't want to delete the
// email from the DB. Instead, we want to just set the 'opt_out'
// to true
func DeleteEmail(db *sql.DB, email string) error {
	_, err := db.Exec(`
		UPDATE email
		SET opt_out=true
		WHERE email=?`, email)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

type GetEmailBatchQueryParams struct {
	Page int // page is for pagination purposes
	Count int // number of email supposed to be returned
}

func GetEmailBatch(db *sql.DB, params GetEmailBatchQueryParams) ([]EmailEntry, error) {
	var empty []EmailEntry

	rows, err := db.Query(`
		SELECT id, email, confirmed_at, opt_out
		FROM emails
		WHERE opt_out = false
		ORDER BY id ASC
		LIMIT ? OFFSET ?`, params.Count, (params.Page-1)*params.Count) // the last line 'LIMIT ? OFFSET ?' is what enables pagination
		
	if err != nil {
		log.Println(err)
		return empty, err
	}

	defer rows.Close()

	emails := make([]EmailEntry, 0, params.Count)

	for rows.Next() {
		email, err := emailEntryFromRow(rows)
		if err != nil {
			return nil, err
		}
		emails = append(emails, *email)
	}

	return emails, nil
}