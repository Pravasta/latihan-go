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
db_host=localhost
db_port=5432
db_user=taskflow
db_password=taskflow
db_name=taskflow
```

These values must match the `postgres` service in `docker-compose.yaml`.

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

The server starts on `http://localhost:8080`.

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
