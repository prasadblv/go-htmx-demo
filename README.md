# ecomm (go-htmx-demo)

A Go web application using server-rendered HTML with [HTMX](https://htmx.org)
and a PostgreSQL data store.

> **Status:** early scaffold. The data-access layer (`models`) is implemented
> and unit-tested; the HTTP server, handlers and middleware are placeholders.

## Architecture

| Package       | Responsibility                                              | Status       |
|---------------|-------------------------------------------------------------|--------------|
| `main`        | Process entry point and start-up wiring                     | Placeholder  |
| `handlers`    | HTTP handlers (authentication endpoints planned)            | Placeholder  |
| `middleware`  | Cross-cutting HTTP concerns (logging, auth, recovery)       | Placeholder  |
| `models`      | `User` entity and parameterized SQL access                  | Implemented  |
| `database`    | PostgreSQL configuration and connection initialization      | Partial      |

## Prerequisites

- Go (see `go.mod` for the required version)
- PostgreSQL with a `users` table, for example:

  ```sql
  CREATE TABLE users (
      id       SERIAL PRIMARY KEY,
      username TEXT NOT NULL UNIQUE,
      password TEXT NOT NULL  -- hashed credential, never plaintext
  );
  ```

## Configuration

Settings are read from environment variables, optionally loaded from a `.env`
file in the working directory. `.env` is git-ignored; do not commit it.

| Variable            | Description                |
|---------------------|----------------------------|
| `POSTGRES_DB`       | Database name              |
| `POSTGRES_USER`     | Database role              |
| `POSTGRES_PASSWORD` | Role password              |
| `POSTGRES_HOST`     | Server host                |
| `POSTGRES_PORT`     | Server port                |

## Testing

```sh
go test ./...
```

Model tests use [go-sqlmock](https://github.com/DATA-DOG/go-sqlmock), so no
running database is required.

## Security notes

- Passwords must be hashed (e.g. bcrypt) before being passed to `models.InsertUser`.
- All queries are parameterized.
- The connection string currently sets `sslmode=disable`; enable TLS in production.

## Known issues

1. `database.InitDatabase` shadows the package-level `Database` handle, so it is never set.
2. No PostgreSQL driver is imported, so `sql.Open("postgres", ...)` fails at runtime.
3. A missing `.env` file is fatal even if the variables exist in the environment.

## License

See [LICENSE](LICENSE).
