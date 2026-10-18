// Package database manages the application's PostgreSQL connection.
//
// Configuration is read from environment variables, optionally populated from
// a .env file in the working directory (see InitDatabase). The .env file is
// excluded from version control and must never be committed.
package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Database is the package-level handle intended to be shared by the
// application once InitDatabase has completed.
//
// Known issue: InitDatabase currently assigns to a locally scoped variable of
// the same name, so this handle is not populated and remains nil. Consumers
// must not rely on it until that defect is corrected.
var Database *sql.DB

// InitDatabase loads configuration, opens a PostgreSQL connection and verifies
// connectivity with a ping.
//
// The following environment variables are required. Values may be supplied by
// the process environment or by a .env file in the working directory:
//
//	POSTGRES_DB        name of the database
//	POSTGRES_USER      role used to authenticate
//	POSTGRES_PASSWORD  password for the role
//	POSTGRES_HOST      server host name or address
//	POSTGRES_PORT      server TCP port
//
// Failure behavior: InitDatabase terminates the process via log.Fatal if the
// .env file cannot be loaded, the connection cannot be opened, or the ping
// fails. It is therefore intended to be called once, during start-up, and not
// from request-handling code. Note that a missing .env file is treated as
// fatal even when the variables are already present in the environment.
//
// Security considerations: the connection string is built with sslmode=disable,
// so traffic is not encrypted in transit. This is acceptable only for local
// development; production deployments should enable TLS (for example
// sslmode=verify-full).
//
// Known issues:
//   - The "postgres" driver is opened by name but no driver package (for
//     example github.com/lib/pq or github.com/jackc/pgx/v5/stdlib) is imported,
//     so sql.Open fails at runtime with an unknown driver error.
//   - The result of sql.Open is assigned with := and shadows the package-level
//     Database variable (see Database).
func InitDatabase() {
	// Load environment variables for the PostgreSQL connection.
	err := godotenv.Load()
	if err != nil {
		log.Fatal(" Error! loading environment variables.")
	}
	dbName := os.Getenv("POSTGRES_DB")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbHost := os.Getenv("POSTGRES_HOST")
	dbPort := os.Getenv("POSTGRES_PORT")

	// Assemble a libpq-style key/value connection string.
	dbSourceName := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable", dbHost, dbPort, dbUser, dbName, dbPassword)

	// sql.Open validates arguments only; it does not establish a connection.
	Database, err := sql.Open("postgres", dbSourceName)

	if err != nil {
		log.Fatal(" Error while connecting to database !", err)
	}

	// Ping forces a round trip to confirm the server is reachable and the
	// credentials are accepted.
	if err := Database.Ping(); err != nil {
		log.Fatal("Unable to establish conncetion with Database")
	}

	log.Println("Database connection successfull.!")
}
