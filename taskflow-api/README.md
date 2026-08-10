# TaskFlow API

A simple task/project management REST API built with Go's standard `net/http`, backed by PostgreSQL.

## Prerequisites

- Go 1.25+
- Docker (for running PostgreSQL via `docker-compose`)
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI

  ```bash
  brew install golang-migrate
  ```

## 1. Configure environment variables

Copy/create a `.env` file in the project root:

```env
APP_PORT=8080

DB_HOST=localhost
DB_PORT=5432
DB_USER=taskflow
DB_PASSWORD=taskflow
DB_NAME=taskflow

JWT_SECRET=my-secret
JWT_EXPIRE=24h
```

`DB_*` values must match the `postgres` service in `docker-compose.yaml`. All variables are loaded by `internal/config` (via [`godotenv`](https://github.com/joho/godotenv)) and fall back to the defaults above if unset — `.env` is optional in that sense, but you should still set your own `JWT_SECRET` rather than rely on the default outside local dev.

| Variable | Default | Used for |
|---|---|---|
| `APP_PORT` | `8080` | HTTP server listen port |
| `DB_HOST` | `localhost` | Postgres host |
| `DB_PORT` | `5432` | Postgres port |
| `DB_USER` | `taskflow` | Postgres user |
| `DB_PASSWORD` | `taskflow` | Postgres password |
| `DB_NAME` | `taskflow` | Postgres database name |
| `JWT_SECRET` | `my-secret` | HMAC signing secret for auth tokens |
| `JWT_EXPIRE` | `24h` | Token expiry duration |

## 2. Start PostgreSQL

```bash
docker compose up -d postgres
```

This starts a `postgres:16-alpine` container on port `5432`, with data persisted in the `postgres_data` volume.

To stop it later:

```bash
docker compose down
```

## 3. Run database migrations

Migration files live in `db/migrations/`. Using `golang-migrate`:

```bash
migrate -path db/migrations \
  -database "postgres://taskflow:taskflow@localhost:5432/taskflow?sslmode=disable" \
  up
```

Other useful commands:

```bash
# roll back the last migration
migrate -path db/migrations \
  -database "postgres://taskflow:taskflow@localhost:5432/taskflow?sslmode=disable" \
  down 1

# check current migration version
migrate -path db/migrations \
  -database "postgres://taskflow:taskflow@localhost:5432/taskflow?sslmode=disable" \
  version

# create a new migration
migrate create -ext sql -dir db/migrations -seq <migration_name>
```

> If your `.env` values differ from the defaults above, update the connection string accordingly.

## 4. Install Go dependencies

```bash
go mod download
```

## 5. Run the server

```bash
go run cmd/api/main.go
```

The server starts on `http://localhost:$APP_PORT` (`8080` by default). `cmd/api/main.go` gets its config, database connection, and JWT service from a single `bootstrap.NewApp()` call — see `internal/bootstrap/app.go` and `internal/config/config.go` if you need to wire up another dependency the same way.

## Running tests

```bash
go test ./...
```

## API overview

All endpoints except `/register` and `/login` require an `Authorization: Bearer <token>` header (obtained from `/login`).

| Method | Path | Description |
|---|---|---|
| POST | `/register` | Create a new user |
| POST | `/login` | Authenticate and receive a JWT |
| GET | `/me` | Get the current authenticated user |
| POST | `/projects` | Create a project |
| GET | `/projects` | List projects owned by the current user |
| GET | `/projects/{id}` | Get a project by ID |
| PUT | `/projects/{id}` | Update a project |
| DELETE | `/projects/{id}` | Delete a project |
| POST | `/projects/{projectID}/tasks` | Create a task under a project |
| GET | `/projects/{projectID}/tasks` | List tasks (supports `limit`, `page`, `status`, `search`, `sort`, `order`) |
| GET | `/projects/{projectID}/tasks/{taskID}` | Get a task by ID |
| PUT | `/projects/{projectID}/tasks/{taskID}` | Update a task's title/description |
| PATCH | `/projects/{projectID}/tasks/{taskID}/status` | Update a task's status |
| DELETE | `/projects/{projectID}/tasks/{taskID}` | Delete a task |
