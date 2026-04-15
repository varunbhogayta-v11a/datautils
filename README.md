# DataUtil - Dataset Processing CLI Tool

A powerful Command-Line Interface tool for processing datasets with filtering, transformation, validation, import/export, and full CRUD database operations with role-based access control.

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                DataUtil CLI                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                        Commands Layer (cmd/)                         │    │
│  │                                                                       │    │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌──────────┐ ┌─────────────┐ │    │
│  │  │ Filter  │ │Transform│ │Validate │ │  Import  │ │ CRUD Ops   │ │    │
│  │  └────┬────┘ └────┬────┘ └────┬────┘ └─────┬────┘ └──────┬──────┘ │    │
│  │       │         │          │            │             │        │    │
│  │  ┌────┴────┐┌───┴────┐┌───┴────┐ ┌────┴─────┐ ┌──────┴──────┐  │    │
│  │  │Query/Sel│ │Insert │ │ Update │ │ Delete   │ │ Export/CSV │  │    │
│  │  └─────────┘ └───────┘ └────────┘ └──────────┘ └────────────┘  │    │
│  └──────────────────────────┬──────────────────────────────────────┘    │
│                              │                                              │
│  ┌──────────────────────────┴──────────────────────────────────────┐    │
│  │                    Auth & Middleware Layer                        │    │
│  │  ┌────────────────┐  ┌──────────────────┐  ┌────────────────┐   │    │
│  │  │ JWT Token Auth  │  │ Role-Based Access │  │  Operations   │   │    │
│  │  │   (login/reg)  │  │   (admin/user)    │  │     Log       │   │    │
│  │  └───────┬────────┘  └────────┬───────────┘  └───────┬────────┘   │    │
│  └──────────┼───────────────────┼──────────────────────┼────────────┘    │
│             │                   │                      │                 │
│  ┌──────────┴───────────────────┴──────────────────────┴────────────┐    │
│  │                         Data Layer (pkg/)                          │    │
│  │                                                                      │    │
│  │  ┌─────────────────┐  ┌─────────────────┐  ┌──────────────────┐   │    │
│  │  │    pkg/data     │  │   pkg/models    │  │   pkg/auth       │   │    │
│  │  │  CSV/JSON/XML   │  │   User/Role     │  │   JWT/JWT        │   │    │
│  │  │  Excel Reader   │  │   Permissions   │  │   Hash/Verify    │   │    │
│  │  └────────┬────────┘  └────────┬────────┘  └────────┬───────────┘   │    │
│  │           │                  │                    │               │    │
│  │  ┌────────┴──────────────────┴────────────────────┴───────────┐   │    │
│  │  │                      pkg/db                              │   │    │
│  │  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐   │   │    │
│  │  │  │   SQLite    │  │ PostgreSQL  │  │      MySQL      │   │   │    │
│  │  │  │  (default)  │  │  (pgx/v5)   │  │                 │   │   │    │
│  │  │  └─────────────┘  └─────────────┘  └─────────────────┘   │   │    │
│  │  └───────────────────────────────────────────────────────────┘   │    │
│  └───────────────────────────────────────────────────────────────────┘    │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Project Structure

```
datautil/
├── main.go                    # Entry point
├── cmd/                       # CLI commands
│   ├── root.go              # Root command & config
│   ├── auth.go              # init-db, register, login, token handling
│   ├── filter.go            # Filter rows/columns from datasets
│   ├── transform.go         # Add/remove/rename columns
│   ├── validate.go          # Schema validation
│   ├── export.go            # Export to different formats
│   ├── import.go            # CSV import to database
│   ├── query.go             # SQL SELECT queries
│   ├── crud.go              # Insert/Update/Delete operations
│   ├── users.go             # List users, view logs
│   └── interactive.go       # Interactive mode
├── pkg/
│   ├── data/
│   │   └── reader.go        # File readers (CSV, JSON, XML, Excel)
│   ├── models/
│   │   └── user.go          # User, Role, OperationLog models
│   ├── auth/
│   │   └── jwt.go          # JWT token generation & validation
│   └── db/
│       └── config.go        # Multi-database configuration
├── tests/
│   └── test_data.*          # Sample test files
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

---

## Features

### 1. Data Operations
| Command | Description |
|---------|-------------|
| **filter** | Filter rows by conditions, select specific columns |
| **transform** | Add, remove, rename columns in datasets |
| **validate** | Validate schema, check required columns & types |
| **export** | Convert between CSV, JSON, XML formats |

### 2. Database CRUD Operations
| Command | Description |
|---------|-------------|
| **query** | Execute SELECT queries |
| **insert** | Insert rows into table |
| **update** | Update rows in table |
| **delete** | Delete rows from table |
| **import** | Import CSV data to database |

### 3. User Management
| Command | Description |
|---------|-------------|
| **init-db** | Initialize database & create tables |
| **register** | Register new user account |
| **login** | Get JWT authentication token |
| **users** | List all users (admin only) |
| **logs** | View operation history |

### 4. Authentication & Security
- JWT-based authentication
- Role-based access control (RBAC)
- Password hashing with bcrypt
- Operation logging for audit

---

## Role-Based Access Control

```
                    ┌─────────────────┐
                    │   User Roles    │
                    └────────┬────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
        ▼                    ▼                    ▼
┌───────────────┐   ┌─────────────────┐   ┌───────────────┐
│     ADMIN    │   │      USER      │   │     GUEST    │
├───────────────┤   ├─────────────────┤   ├───────────────┤
│ ✓ Read        │   │ ✓ Read          │   │ ✓ Read        │
│ ✓ Write       │   │ ✓ Write         │   │ ✗ Write       │
│ ✓ Delete      │   │ ✗ Delete        │   │ ✗ Delete      │
│ ✓ Admin       │   │ ✗ Admin         │   │ ✗ Admin       │
└───────────────┘   └─────────────────┘   └───────────────┘
```

### Permissions by Role
| Action | Admin | User | Guest |
|--------|-------|------|-------|
| filter/transform/validate | ✓ | ✓ | ✗ |
| query/select | ✓ | ✓ | ✓ |
| insert | ✓ | ✓ | ✗ |
| update | ✓ | ✓ | ✗ |
| delete | ✓ | ✗ | ✗ |
| user management | ✓ | ✗ | ✗ |

---

## Supported Databases

| Driver | Environment Variables |
|--------|----------------------|
| **SQLite** (default) | `DB_DRIVER=sqlite3` |
| **PostgreSQL** | `DB_DRIVER=postgres` |
| **MySQL** | `DB_DRIVER=mysql` |

### Database Connection Config
```bash
export DB_HOST=localhost       # Database host
export DB_PORT=5432            # Database port
export DB_USER=postgres        # Database user
export DB_PASSWORD=secret      # Database password
export DB_NAME=datautil        # Database name
export DB_DRIVER=sqlite3       # sqlite3, postgres, mysql
```

---

## Quick Start Guide

### 1. Initialize Database
```bash
# SQLite (default)
./datautil init-db

# PostgreSQL
export DB_DRIVER=postgres DB_HOST=localhost DB_USER=postgres DB_PASSWORD=secret DB_NAME=mydb
./datautil init-db
```

### 2. Register & Login
```bash
# Register new user
./datautil register --username admin --email admin@example.com --password secret123

# Login to get token
./datautil login --email admin@example.com --password secret123

# Output: JWT Token will be displayed
# Use this token with --token flag
```

### 3. Data Operations (with file-based datasets)
```bash
TOKEN="your-jwt-token-here"

# Filter data
./datautil filter --input data.csv --where "age > 25" --token "$TOKEN"

# Transform data  
./datautil transform --input data.csv --add "full_name=first_name+last_name" --token "$TOKEN"

# Validate data
./datautil validate --input data.csv --required name,email,age --token "$TOKEN"

# Export data
./datautil export --input data.csv --to json --output result.json --token "$TOKEN"
```

### 4. Database CRUD Operations
```bash
TOKEN="your-jwt-token-here"

# Query data
./datautil query --sql "SELECT * FROM users" --token "$TOKEN"

# Insert data
./datautil insert --table users --values "name=John,age=30,city=NYC" --token "$TOKEN"

# Update data
./datautil update --table users --set "age=31" --where "name=John" --token "$TOKEN"

# Delete data
./datautil delete --table users --where "name=John" --token "$TOKEN"

# Import CSV to database
./datautil import --input data.csv --table my_data --create --token "$TOKEN"
```

---

## Usage Examples

### Filter Operations
```bash
# Filter rows by condition
./datautil filter --input data.csv --where "age > 25" --token "$TOKEN"

# Select specific columns
./datautil filter --input data.csv --select name,email --token "$TOKEN"

# Combine filter and column selection
./datautil filter --input data.csv --where "city==NYC" --select name,age --token "$TOKEN"
```

### Transform Operations
```bash
# Add new column
./datautil transform --input data.csv --add "full_name=first+last" --token "$TOKEN"

# Remove columns
./datautil transform --input data.csv --remove age,city --token "$TOKEN"

# Rename column
./datautil transform --input data.csv --rename "old_name:new_name" --token "$TOKEN"
```

### Database Query
```bash
# Simple query
./datautil query --sql "SELECT * FROM users" --token "$TOKEN"

# Query with condition
./datautil query --sql "SELECT * FROM products WHERE price > 100" --token "$TOKEN"

# Aggregate query
./datautil query --sql "SELECT COUNT(*) as total FROM orders" --token "$TOKEN"
```

---

## Environment Configuration

```bash
# Create .env file
cat > .env << 'EOF'
DB_DRIVER=sqlite3
DB_NAME=datautil
JWT_SECRET=your-secret-key-change-in-production
EOF

# Or use environment variables
export DB_DRIVER=postgres
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=secret
export DB_NAME=myapp
```

---

## Docker/Podman Support

```bash
# Build container
podman build -t datautil .

# Run with volume mount
podman run --rm -v $(pwd):/data datautil filter --input /data/data.csv --where "age > 25" --token "JWT"
```

---

## Version

1.0.0

## License

MIT License - Improwised Technologies Pvt. Ltd.