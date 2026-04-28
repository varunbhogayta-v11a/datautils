# DataUtil - CLI Tool for Automated Dataset Processing

## Overview

DataUtil is a Go-based CLI tool for dataset processing (filtering, transforming, validating, exporting). It supports CSV, JSON, XML, and Excel formats with PostgreSQL/SQLite database persistence and JWT-based authentication.

---

## Project Structure

```
datautil/
├── cmd/              - CLI commands (12 files)
├── pkg/              - Core packages
│   ├── auth/         - JWT authentication
│   ├── data/         - Dataset reader
│   ├── db/           - Database config
│   ├── models/       - Data models
│   ├── operations/    - Filter operations
│   └── repo/          - Repository
└── tests/            - Test files
```

---

## Key Features

- **CLI Commands**: filter, transform, validate, export, import, query, users, CRUD operations
- **HTTP API Server**: RESTful endpoints with Swagger UI
- **Authentication**: JWT-based auth with register/login
- **File Formats**: CSV, JSON, XML, Excel
- **Database**: PostgreSQL & SQLite via GORM

---

## Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| spf13/cobra | 1.8.0 | CLI framework |
| spf13/viper | 1.18.2 | Config management |
| gorm.io/gorm | 1.31.1 | ORM |
| gorm.io/driver/postgres | 1.6.0 | PostgreSQL driver |
| gorm.io/driver/sqlite | 1.6.0 | SQLite driver |
| golang-jwt/jwt/v5 | 5.3.1 | JWT auth |
| xuri/excelize/v2 | 2.8.0 | Excel handling |

---

## CLI Commands

| Command | Description | Example |
|---------|-------------|---------|
| `filter` | Filter dataset rows | `./datautil filter --input data.csv --where "age > 25"` |
| `transform` | Transform dataset | `./datautil transform --input data.csv --add city=NYC` |
| `validate` | Validate dataset | `./datautil validate --input data.csv --required name,age` |
| `export` | Export to format | `./datautil export --input data.csv --to json` |
| `import` | Import to database | `./datautil import --source data.csv --to-db` |
| `query` | Execute SQL | `./datautil query --sql "SELECT * FROM users"` |
| `register` | Register user | `./datautil register --username admin --email admin@example.com --password pass` |
| `login` | Login user | `./datautil login --email admin@example.com --password pass` |
| `init-db` | Initialize database | `./datautil init-db` |
| `server` | Start API server | `./datautil server --port 8080` |

---

## HTTP API Endpoints

### Public Endpoints (No Auth Required)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/health` | GET | Health check |
| `/api/auth/register` | POST | Register new user |
| `/api/auth/login` | POST | Login and get JWT |

### Protected Endpoints (Auth Required)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/data/filter` | POST | Filter rows |
| `/api/data/transform` | POST | Transform data |
| `/api/data/validate` | POST | Validate dataset |
| `/api/data/export` | POST | Export data |
| `/api/data/import` | POST | Import to DB |
| `/api/db/query` | POST | Execute SQL |
| `/api/db/insert` | POST | Insert row |
| `/api/db/update` | POST | Update rows |
| `/api/db/delete` | POST | Delete rows |
| `/api/users` | GET | List users |
| `/api/logs` | GET | Operation logs |

### Swagger UI

| URL | Description |
|-----|-------------|
| `/swagger` | Interactive documentation |
| `/swagger.yaml` | OpenAPI spec |

---

## Data Models

### User
```go
type User struct {
    ID        uint
    Username  string
    Email     string
    Password  string
    Role      string    // "admin", "user", "guest"
    Active    bool
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### OperationLog
```go
type OperationLog struct {
    ID         uint
    UserID     uint
    Operation  string
    InputFile  string
    OutputFile string
    Details    string
    CreatedAt  time.Time
}
```

### RouteConfig
```go
type RouteConfig struct {
    ID                uint
    Path              string
    SourceType        string
    Source            string
    FilterExpr        string
    SelectCols        string
    AuthRequired      bool
    IsDynamic         bool
    PaginationEnabled bool
    DefaultLimit      int
    MaxLimit          int
}
```

---

## Testing Coverage

| Package | Coverage | Tests |
|---------|----------|-------|
| `pkg/operations` | 92.6% | 21 |
| `pkg/models` | 80.0% | 9 |
| `pkg/auth` | 83.1% | 21 |
| `pkg/data` | 79.7% | 16 |
| `pkg/db` | 9.2% | 8 |
| **Total** | **~67%** | **75+** |

---

## Build Commands

```bash
# Build binary
go build -o datautil

# Run tests
go test ./...

# Run with coverage
go test -cover ./...

# Run integration tests
go test -tags=integration ./tests/...

# Run E2E tests
go test -tags=e2e ./tests/...
```

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | localhost | Database host |
| `DB_PORT` | 5432 | Database port |
| `DB_USER` | postgres | Database user |
| `DB_PASSWORD` | postgres | Database password |
| `DB_NAME` | datautil | Database name |
| `DB_DRIVER` | sqlite3 | Database driver |
| `JWT_SECRET` | - | JWT signing secret |
| `PORT` | 8080 | HTTP server port |

---

## API Usage Example

```bash
# Start server
./datautil server &

# Register user
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","email":"admin@example.com","password":"password123"}'

# Login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password123"}'

# Filter data (use token from login)
curl -X POST http://localhost:8080/api/data/filter \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"input":"data.csv","where":"age > 25"}'
```