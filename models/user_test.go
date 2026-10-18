package models

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// Package-level SQL matchers. sqlmock treats expectations as regular
// expressions, so parentheses and placeholders are escaped. These must be kept
// in sync with the statements in user.go.
const (
	selectByName = "SELECT id, username, password FROM users WHERE username = \\$1"
	selectByID   = "SELECT id, username, password FROM users WHERE id=\\$1"
	insertUser   = "INSERT INTO users \\(username, password\\) VALUES \\(\\$1, \\$2\\)"
)

// newMock returns a sqlmock-backed *sql.DB together with its expectation
// controller. The connection is closed automatically when the test completes.
func newMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, mock
}

// TestGetUserbyName verifies GetUserbyName for a successful lookup, a missing
// row (sql.ErrNoRows), a driver error and a row-scan failure.
func TestGetUserbyName(t *testing.T) {
	dbErr := errors.New("connection refused")

	tests := []struct {
		name     string
		username string
		setup    func(m sqlmock.Sqlmock)
		wantUser *User
		wantErr  error
	}{
		{
			name:     "found",
			username: "alice",
			setup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(selectByName).WithArgs("alice").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password"}).AddRow(1, "alice", "hash"))
			},
			wantUser: &User{ID: 1, Username: "alice", Password: "hash"},
		},
		{
			name:     "not found returns ErrNoRows",
			username: "ghost",
			setup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(selectByName).WithArgs("ghost").WillReturnError(sql.ErrNoRows)
			},
			wantErr: sql.ErrNoRows,
		},
		{
			name:     "database error is propagated",
			username: "alice",
			setup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(selectByName).WithArgs("alice").WillReturnError(dbErr)
			},
			wantErr: dbErr,
		},
		{
			name:     "scan error is propagated",
			username: "alice",
			setup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(selectByName).WithArgs("alice").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password"}).AddRow("notint", "alice", "hash"))
			},
			wantErr: errors.New("any"), // any non-nil error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := newMock(t)
			tt.setup(mock)

			got, err := GetUserbyName(db, tt.username)

			checkResult(t, got, err, tt.wantUser, tt.wantErr)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet expectations: %v", err)
			}
		})
	}
}

// TestGetUserByID verifies GetUserByID for a successful lookup, a missing row
// (sql.ErrNoRows) and a propagated driver error.
func TestGetUserByID(t *testing.T) {
	dbErr := errors.New("connection refused")

	tests := []struct {
		name     string
		id       int
		setup    func(m sqlmock.Sqlmock)
		wantUser *User
		wantErr  error
	}{
		{
			name: "found",
			id:   7,
			setup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(selectByID).WithArgs(7).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password"}).AddRow(7, "bob", "hash"))
			},
			wantUser: &User{ID: 7, Username: "bob", Password: "hash"},
		},
		{
			name: "not found returns ErrNoRows",
			id:   99,
			setup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(selectByID).WithArgs(99).WillReturnError(sql.ErrNoRows)
			},
			wantErr: sql.ErrNoRows,
		},
		{
			name: "database error is propagated",
			id:   7,
			setup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(selectByID).WithArgs(7).WillReturnError(dbErr)
			},
			wantErr: dbErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := newMock(t)
			tt.setup(mock)

			got, err := GetUserByID(db, tt.id)

			checkResult(t, got, err, tt.wantUser, tt.wantErr)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet expectations: %v", err)
			}
		})
	}
}

// TestInsertUser verifies that InsertUser stores the hash as a string and
// propagates execution errors such as unique-constraint violations.
func TestInsertUser(t *testing.T) {
	t.Run("success stores hashed password as string", func(t *testing.T) {
		db, mock := newMock(t)
		mock.ExpectExec(insertUser).WithArgs("alice", "hashed").WillReturnResult(sqlmock.NewResult(1, 1))

		if err := InsertUser(db, "alice", []byte("hashed")); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("exec error is propagated", func(t *testing.T) {
		db, mock := newMock(t)
		dbErr := errors.New("duplicate key value violates unique constraint")
		mock.ExpectExec(insertUser).WithArgs("alice", "hashed").WillReturnError(dbErr)

		if err := InsertUser(db, "alice", []byte("hashed")); !errors.Is(err, dbErr) {
			t.Fatalf("err = %v, want %v", err, dbErr)
		}
	})
}

// checkResult asserts either a matching user (no error) or a non-nil error
// (matching wantErr via errors.Is, unless wantErr is a generic sentinel) with a nil user.
func checkResult(t *testing.T, got *User, err error, wantUser *User, wantErr error) {
	t.Helper()
	if wantErr != nil {
		if err == nil {
			t.Fatalf("expected error, got nil (user=%+v)", got)
		}
		if (wantErr == sql.ErrNoRows || wantErr.Error() != "any") && !errors.Is(err, wantErr) {
			t.Errorf("err = %v, want %v", err, wantErr)
		}
		if got != nil {
			t.Errorf("user = %+v, want nil", got)
		}
		return
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || *got != *wantUser {
		t.Errorf("user = %+v, want %+v", got, wantUser)
	}
}
