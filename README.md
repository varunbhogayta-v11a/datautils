# DataUtil - CLI Tool for Dataset Processing

## Overview

DataUtil is a Go-based CLI tool and API server for dataset processing. It supports filtering, transforming, validating, and exporting data in multiple formats.

**Supported formats:** CSV, JSON, XML, Excel

**Database:** SQLite (built-in, no setup required)

---

## Quick Start

```bash
# 1. Build the binary
go build -o datautil

# 2. Initialize database (creates datautil.db)
./datautil init-db

# 3. Register a user
./datautil register --username demo --email demo@test.com --password demo123

# 4. Start API server
./datautil server
```

Server runs at: http://localhost:8080

---

## Features

| Feature | CLI | API |
|---------|-----|-----|
| Filter data | ✅ | ✅ |
| Transform data | ✅ | ✅ |
| Validate data | ✅ | ✅ |
| Export format | ✅ | ✅ |
| Upload files | - | ✅ |
| Query DB | ✅ | ✅ |
| JWT Auth | ✅ | ✅ |
| Swagger UI | - | ✅ |

---

## CLI Commands

```bash
./datautil init-db              # Initialize SQLite database
./datautil register --username demo --email demo@test.com --password pass123
./datautil login --email demo@test.com --password pass123
./datautil filter --input data.csv --where "age > 25"
./datautil transform --input data.csv --add full_name=name
./datautil validate --input data.csv --required name,age
./datautil export --input data.csv --to json --output out.json
./datautil query --sql "SELECT * FROM users"
./datautil server              # Start API server
```

---

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/health` | GET | Health check |
| `/api/auth/register` | POST | Register user |
| `/api/auth/login` | POST | Login (get token) |
| `/api/data/upload` | POST | Upload file |
| `/api/data/files` | GET | List files |
| `/api/data/download/:file` | GET | Download file |
| `/api/data/filter` | POST | Filter data |
| `/api/data/transform` | POST | Transform data |
| `/api/data/validate` | POST | Validate data |
| `/api/data/export` | POST | Export data |
| `/swagger` | - | Interactive API docs |

---

## API Usage

```bash
# Register
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"demo","email":"demo@test.com","password":"demo123"}'

# Login (get token)
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@test.com","password":"demo123"}'

# Upload file
curl -X POST http://localhost:8080/api/data/upload \
  -H "Authorization: Bearer TOKEN" \
  -F "file=@data.csv"

# List files
curl -X GET http://localhost:8080/api/data/files \
  -H "Authorization: Bearer TOKEN"

# Filter data
curl -X POST http://localhost:8080/api/data/filter \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer TOKEN" \
  -d '{"input":"data/uploads/data.csv","where":"age > 25"}'

# Download result
curl -X GET http://localhost:8080/api/data/download/output.csv \
  -H "Authorization: Bearer TOKEN"
```

---

## Project Structure

```
datautil/
├── cmd/              - CLI commands
├── pkg/              - Core packages
│   ├── auth/         - JWT authentication
│   ├── data/         - Dataset reader
│   ├── db/           - Database config
│   ├── models/       - Data models
│   ├── operations/   - Filter operations
│   └── repo/         - Repository
├── data/             - Uploaded files
│   └── uploads/      - User uploads
├── tests/            - Test data files
├── datautil.db       - SQLite database
└── datautil         - Binary
```

---

## Build for Different Platforms

```bash
# macOS arm64 (M1/M2/M3)
GOOS=darwin GOARCH=arm64 go build -o datautil

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o datautil

# Linux
GOOS=linux GOARCH=amd64 go build -o datautil

# Windows
GOOS=windows GOARCH=amd64 go build -o datautil.exe
```

---

## Testing

```bash
go test ./...
go test -cover ./...
```

---

## Swagger UI

Open http://localhost:8080/swagger for interactive API testing.

---

## Files

- `datautil.db` - SQLite database (auto-created)
- `data/uploads/` - User uploaded files
- `swagger.yaml` - OpenAPI spec (served at `/swagger.yaml`)