# AGENTS.md - Developer Guide for datautil

## Project Overview

datautil is a Go-based CLI tool for dataset processing (filtering, transforming, validating, exporting). It supports CSV, JSON, XML, and Excel formats, with PostgreSQL/SQLite for persistence and JWT-based authentication.

## Build Commands

```bash
# Build the binary
go build -o datautil

# Run all tests
go test ./...

# Run a single test
go test -run TestFilterRows ./pkg/operations/

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Format code
go fmt ./...

# Run go vet
go vet ./...

# Run linter (if golangci-lint is installed)
golangci-lint run
```

## Code Style Guidelines

### Imports

- Group imports: standard library first, then external packages
- Use canonical import paths: `github.com/improwised/datautil/pkg/...`

### Formatting

- Use `go fmt` for formatting (enforces Go standards)
- 4-space indentation (standard Go)
- Max line length: ~100 characters (avoid exceeding 120)

### Types

- Use explicit types; avoid `var x` without type
- Use pointers for large structs or when nil is meaningful
- Prefer interfaces for dependencies (e.g., `io.Reader`, `io.Writer`)

### Naming Conventions

- **Variables/Functions**: camelCase (`filterRows`, `userName`)
- **Constants**: PascalCase or camelCase with prefix (`const MaxRetries = 3`)
- **Types/Interfaces**: PascalCase (`Dataset`, `ValidationResult`)
- **Packages**: lowercase, short (`operations`, `auth`, `models`)
- **Files**: lowercase with underscores (`filter.go`, `auth_jwt.go`)

### Error Handling

- Return errors as last return value: `func() (Result, error)`
- Use `fmt.Errorf` with context: `fmt.Errorf("failed to filter: %w", err)`
- Handle errors explicitly; avoid silent failures
- Wrap errors with context in CLI commands

### Testing

- Test files: `*_test.go` in same package as implementation
- Use table-driven tests for multiple test cases
- Test naming: `Test<FunctionName>_<Scenario>`
- Use `t.Fatalf` for unexpected errors, `t.Errorf` for assertion failures

### Project Structure

```
cmd/          - CLI command implementations
pkg/
  auth/       - JWT authentication
  data/       - Dataset structures and helpers
  db/         - Database connections and migrations
  models/     - Data models
  operations/ - Core data operations (filter, transform, validate)
  repo/       - Data repositories
  utils/      - Utility functions
configs/      - Configuration files
tests/        - Test data files
```

### CLI Commands

- Use Cobra for CLI structure
- Commands in `cmd/` package
- Flags via `PersistentFlags` for global, `Flags` for command-specific

### Database

- Uses GORM for ORM
- Supports PostgreSQL (primary) and SQLite
- Configure via `.env` or `config.yaml`

### Configuration

- Viper for config management
- Config search order: explicit path, `./config.yaml`, `./configs/`, `$HOME/.datautil`
- Environment variables override config file values

## Running Tests

This project uses the standard Go testing package. Tests are located in `*_test.go` files alongside the code they test.

### Test Setup Requirements

Before running integration tests, ensure:
1. PostgreSQL is running (or set DB_DRIVER=sqlite in .env)
2. Database is initialized: `./datautil init-db`
3. Test user exists: `./datautil register --username testuser --email test@example.com --password test123`
4. JWT token is obtained: `./datautil login --email test@example.com --password test123`

A helper script `test_setup.sh` automates this setup process.

### Test Data Files

Place test data files in the `tests/` directory. Example test CSV format:
```csv
name,age,city
John,25,NYC
Jane,30,LA
```

### Test Environment Variables

Create a `.env` file in the project root:
```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=datautil_test
JWT_SECRET=test-secret-key-for-testing
```

## Database Setup

### Supported Databases

- PostgreSQL (primary, recommended)
- SQLite (for development/testing)

### Database Initialization

```bash
# Using PostgreSQL (default)
./datautil init-db

# Using SQLite
DB_DRIVER=sqlite ./datautil init-db
```

### Database Migrations

Database schema is managed through GORM automigration. On first run, tables are created automatically.

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| DB_HOST | Database host | localhost |
| DB_PORT | Database port | 5432 |
| DB_USER | Database user | postgres |
| DB_PASSWORD | Database password | postgres |
| DB_NAME | Database name | datautil |
| DB_DRIVER | Database driver | postgres |
| JWT_SECRET | JWT signing secret | - |
| JWT_EXPIRY | Token expiry duration | 24h |
| PORT | HTTP server port | 8080 |

## CLI Commands

The CLI uses Cobra. Available commands:

```bash
# Authentication
./datautil register --username <name> --email <email> --password <pass>
./datautil login --email <email> --password <pass>

# Data operations
./datautil filter --input <file> --where "<condition>"
./datautil transform --input <file> --add "<column>=<expression>"
./datautil validate --input <file> --required <columns>
./datautil export --input <file> --to <format> --output <file>

# Database operations
./datautil import --source <file> --to-db
./datautil query --sql "<query>"
./datautil users --token <jwt>

# Administration
./datautil init-db
./datautil logs --token <jwt>
```

## Error Codes

The application uses standard Go error handling. All functions return `(result, error)` where error is:
- `nil` on success
- An error wrapped with `fmt.Errorf` on failure

Common error patterns:
- File not found: `fmt.Errorf("file not found: %s", path)`
- Invalid input: `fmt.Errorf("invalid %s: %w", field, err)`
- Database errors: `fmt.Errorf("database error: %w", err)`

## Dependencies

Key external packages:
- `github.com/spf13/cobra` - CLI command framework
- `github.com/spf13/viper` - Configuration management
- `gorm.io/gorm` - ORM
- `gorm.io/driver/postgres` - PostgreSQL driver
- `gorm.io/driver/sqlite` - SQLite driver
- `github.com/golang-jwt/jwt/v5` - JWT authentication
- `github.com/xuri/excelize/v2` - Excel file handling

## Development Workflow

1. Write code following Go conventions
2. Run `go fmt` to format code
3. Run `go vet` to check for issues
4. Run tests: `go test ./...`
5. Build binary: `go build -o datautil`
6. Test CLI commands manually

## HTTP API Server

The project includes an HTTP API server with interactive Swagger documentation.

### Starting the Server

```bash
# Start the API server (default port 8080)
./datautil server

# Start on custom port
./datautil server --port 3000

# Start on custom host
./datautil server --host 127.0.0.1 --port 8080
```

### API Endpoints

| Endpoint | Method | Description | Auth Required |
|----------|--------|-------------|----------------|
| `/api/health` | GET | Health check | No |
| `/api/auth/register` | POST | Register new user | No |
| `/api/auth/login` | POST | Login and get JWT | No |
| `/api/data/filter` | POST | Filter dataset rows | Yes |
| `/api/data/transform` | POST | Transform dataset | Yes |
| `/api/data/validate` | POST | Validate dataset | Yes |
| `/api/data/export` | POST | Export dataset | Yes |
| `/api/data/import` | POST | Import file to DB | Yes |
| `/api/db/query` | POST | Execute SQL query | Yes |
| `/api/db/insert` | POST | Insert into table | Yes |
| `/api/db/update` | POST | Update table rows | Yes |
| `/api/db/delete` | POST | Delete table rows | Yes |
| `/api/users` | GET | List all users | Yes |
| `/api/logs` | GET | Get operation logs | Yes |

### Swagger Documentation

Access interactive Swagger UI at:
- `/swagger` - Swagger UI interface
- `/swagger.yaml` - OpenAPI specification (raw YAML)

### API Usage Example

```bash
# Start server
./datautil server &

# Register a user (note the token in response)
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","email":"admin@example.com","password":"password123"}'

# Login to get token
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password123"}'

# Filter data (use token from login)
curl -X POST http://localhost:8080/api/data/filter \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"input":"tests/test_data.csv","where":"age > 25"}'
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| PORT | HTTP server port | 8080 |
| HOST | HTTP server host | 0.0.0.0 |
| DB_HOST | Database host | localhost |
| DB_PORT | Database port | 5432 |
| DB_USER | Database user | postgres |
| DB_PASSWORD | Database password | postgres |
| DB_NAME | Database name | datautil |
| DB_DRIVER | Database driver | sqlite3 |
| JWT_SECRET | JWT signing secret | datautil-secret-key-change-in-production |
| JWT_EXPIRY | Token expiry duration | 24h |