# Repository Guidelines

## Project Structure & Module Organization

- `cmd/api/`: API entry point.
- `internal/server/`: HTTP routes, handlers, helpers, and handler tests.
- `internal/database/`: database connection service and integration tests.
- `internal/db/migrations/`: Goose migrations embedded by `fs.go`.
- `internal/db/queries/`: SQLC query definitions.
- `internal/db/sqlc/`: generated SQLC Go code; do not edit manually.
- `frontend/src/`: React application source.
- `frontend/public/`: static frontend assets.

## Build, Test, and Development Commands

- `make build`: builds the Go API binary as `main`.
- `make test`: runs `go test ./... -v`.
- `make all`: runs build and test.
- `make run`: starts the Go API and the Vite dev server.
- `make docker-run`: starts services with Docker Compose.
- `make docker-down`: stops Docker Compose services.
- `make itest`: runs database integration tests under `internal/database`.
- `make watch`: starts live reload with `air`.
- `npm run build --prefix frontend`: type-checks and builds the React app.
- `npm run lint --prefix frontend`: runs ESLint for frontend code.

If Go cannot write to the default build cache in restricted environments, use `GOCACHE=/tmp/go-build-cache go test ./...`.

## Coding Style & Naming Conventions

Format Go code with `gofmt`. Keep handlers grouped by resource, for example `tenants.go` and `brokers.go`. Name exported handlers with verb-resource patterns such as `CreateBroker`, `ListTenants`, and `UpdateBrokerRole`.

Use SQLC for database access. Edit files in `internal/db/queries/` and regenerate SQLC output instead of manually changing `internal/db/sqlc/`.

Frontend code uses TypeScript, React, Vite, and ESLint. Keep component files in `frontend/src/` and use PascalCase component names.

## Testing Guidelines

Go tests use `testing` and `httptest`. Name test files `*_test.go` and tests `Test...`, for example `TestCreateBroker`. Prefer `httptest.NewRequest` and `httptest.ResponseRecorder` for handler tests.

Run `make test` before opening a pull request. Run `make itest` when changing database connection logic, migrations, or SQL queries.

## Commit & Pull Request Guidelines

Recent commits use short past-tense summaries, such as `added brokers handlers` and `fixed the schema`. Keep commits focused and describe the behavior changed.

Pull requests should include a brief summary, test commands run, and any database migration or frontend impact. Include screenshots for visible frontend changes. Link related issues when available.

## Security & Configuration Tips

Do not commit secrets or local `.env` values. Configuration is loaded with `godotenv`, and local services are expected to run through Docker Compose during development.
