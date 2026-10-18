package models

import (
	"database/sql"
	"log"
)

// User represents a row in the users table.
type User struct {
	// ID is the database-generated primary key.
	ID int
	// Username is the unique login name.
	Username string
	// Password holds the hashed credential (never plaintext). It is populated
	// by the lookup functions, so callers must not serialize or log a User
	// without first clearing this field.
	Password string
}

// GetUserbyName retrieves the user whose username exactly matches the
// supplied value (case-sensitive).
//
// It returns the matching *User on success. If no row exists, it returns
// (nil, sql.ErrNoRows) and logs nothing; callers should test for this with
// errors.Is. Any other failure (connectivity, scan/type mismatch) is logged
// and returned as (nil, err) unchanged.
//
// The query is parameterized, so username is safe to pass from untrusted input.
func GetUserbyName(db *sql.DB, username string) (*User, error) {
	query := "SELECT id, username, password FROM users WHERE username = $1"
	user := &User{}
	err := db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("Error while fetching values from database! ")
		}
		return nil, err
	}
	return user, nil
}

// InsertUser persists a new user with the given username and password hash.
//
// hashedPassword must already be hashed with a suitable algorithm (for example
// bcrypt); it is stored verbatim as a string and is never hashed here. Passing
// a plaintext password is a security defect.
//
// It returns nil on success. Failures, including violation of the username
// uniqueness constraint, are logged and returned unchanged; callers can
// inspect the error to distinguish duplicates from other faults.
func InsertUser(db *sql.DB, username string, hashedPassword []byte) error {
	query := "INSERT INTO users (username, password) VALUES ($1, $2)"
	_, err := db.Exec(query, username, string(hashedPassword))
	if err != nil {
		log.Println("Error while inserting user data!")
		return err
	}
	return nil
}

// GetUserByID retrieves the user with the given primary key.
//
// Return semantics match GetUserbyName: (nil, sql.ErrNoRows) when no row
// exists, and (nil, err) for any other failure, which is also logged.
func GetUserByID(db *sql.DB, id int) (*User, error) {
	query := "SELECT id, username, password FROM users WHERE id=$1"
	user := &User{}
	err := db.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("Error while fetching values from database! ")
		}
		return nil, err
	}
	return user, nil
}
