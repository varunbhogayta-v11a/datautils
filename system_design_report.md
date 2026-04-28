# Chapter 5: System Design & Implementation

## 5.1 System Design & Methodology

### Overview
The DataUtil system follows a **modular, layered architecture** designed for processing, transforming, and validating datasets across multiple formats (CSV, JSON, XML, Excel). The system employs a **microservices-friendly** design pattern with clear separation of concerns between CLI operations, HTTP API services, and data persistence layers.

### Architecture Pattern
```

                        Client Layer                           
  (CLI Commands / HTTP API Clients / Swagger UI)               

                     
                     

                  Application Layer                           
  • Cobra CLI Framework                                       
  • HTTP Server (RESTful API)                                
  • Authentication & Authorization (JWT)                     
  • Request Validation & Processing                          

                     
                     

                  Business Logic Layer                         
  • Data Operations (Filter, Transform, Validate, Export)    
  • Workflow Orchestration                                   
  • Error Handling & Logging                                 
  • Security Services                                        

                     
                     

                  Data Access Layer                            
  • Repository Pattern                                       
  • GORM ORM (Database Operations)                           
  • File I/O Operations                                      
  • Connection Pooling                                       

                     
                     

                  Data Layer                                  
  • PostgreSQL (Primary Database)                            
  • SQLite (Development/Testing)                             
  • File Storage (CSV, JSON, XML, Excel)                     

```

### Methodology

#### 1. **Domain-Driven Design (DDD)**
- Clear bounded contexts: `auth`, `operations`, `data`, `db`, `models`
- Each package encapsulates specific business capabilities
- Interfaces define contracts between layers

#### 2. **Clean Architecture Principles**
- Dependency rule: Inner layers don't depend on outer layers
- Core business logic independent of frameworks
- Easy to test and maintain

#### 3. **CQRS Pattern (Command Query Responsibility Segregation)**
- Write operations (import, transform, update) separated from read operations (filter, query, export)
- Optimizes performance and scalability

#### 4. **Repository Pattern**
- Abstracts data access logic
- Provides testability with mock implementations
- Decouples business logic from persistence details

### Design Decisions

| Decision | Rationale | Benefit |
|----------|-----------|---------|
| Go Language | Performance, concurrency, static typing | High-throughput data processing |
| Cobra CLI | Industry-standard, extensible | Consistent CLI experience |
| GORM ORM | Supports multiple databases, migrations | Database flexibility |
| JWT Authentication | Stateless, scalable | No server-side session storage |
| Viper Config | Multi-format, environment override | Flexible deployment |
| Layered Architecture | Separation of concerns | Maintainability & testability |

### Mermaid Diagram: System Architecture
```mermaid
graph TB
    subgraph "Client Layer"
        CLI[CLI Commands<br/>cobra-based]
        API[HTTP API<br/>RESTful]
        WEB[Swagger UI<br/>Interactive Docs]
    end
    
    subgraph "Application Layer"
        AUTH[Authentication<br/>JWT Handler]
        VAL[Request Validation<br/>Input Sanitization]
        ROUTER[HTTP Router<br/>Endpoint Management]
        CMD[CLI Command Handlers<br/>Business Logic Calls]
    end
    
    subgraph "Business Logic Layer"
        FILTER[Filter Operations<br/>Row Selection]
        TRANSFORM[Transform Operations<br/>Column Operations]
        VALIDATE[Validation Operations<br/>Data Quality Checks]
        EXPORT[Export Operations<br/>Format Conversion]
        SEC[Security Services<br/>Access Control]
        LOG[Logging Service<br/>Audit Trails]
    end
    
    subgraph "Data Access Layer"
        REPO[Repository Interfaces<br/>Abstract Data Access]
        GORM[GORM ORM<br/>Database Operations]
        FILEIO[File I/O<br/>CSV/JSON/XML/Excel]
        POOL[Connection Pool<br/>DB Resource Management]
    end
    
    subgraph "Data Layer"
        PG[(PostgreSQL<br/>Primary Database)]
        SQLITE[(SQLite<br/>Dev/Test Database)]
        STORAGE[File Storage<br/>Dataset Files]
    end
    
    CLI -->|Commands| CMD
    API -->|HTTP Requests| ROUTER
    WEB -->|API Calls| ROUTER
    
    ROUTER --> AUTH
    ROUTER --> VAL
    ROUTER --> CMD
    
    CMD --> FILTER
    CMD --> TRANSFORM
    CMD --> VALIDATE
    CMD --> EXPORT
    CMD --> SEC
    CMD --> LOG
    
    FILTER --> REPO
    TRANSFORM --> REPO
    VALIDATE --> REPO
    EXPORT --> REPO
    SEC --> REPO
    LOG --> REPO
    
    REPO --> GORM
    REPO --> FILEIO
    
    GORM --> POOL
    POOL --> PG
    POOL --> SQLITE
    
    FILEIO --> STORAGE
```

---

## 5.2 Database Design / Data Structure Design

### Database Schema Design

#### Entity-Relationship Diagram
```mermaid
erDiagram
    USER ||--o{ DATASET : owns
    USER ||--o{ OPERATION_LOG : performs
    USER ||--o{ USER_SESSION : has
    DATASET ||--o{ COLUMN : contains
    DATASET ||--o{ DATASET_VERSION : has
    
    USER {
        bigint id PK
        string username UQ
        string email UQ
        string password_hash
        string full_name
        boolean is_active
        timestamp created_at
        timestamp updated_at
    }
    
    USER_SESSION {
        bigint id PK
        bigint user_id FK
        string token_hash
        string ip_address
        string user_agent
        timestamp expires_at
        timestamp created_at
    }
    
    DATASET {
        bigint id PK
        bigint user_id FK
        string name
        string description
        string original_filename
        string format
        string source_type
        long file_size
        long row_count
        timestamp uploaded_at
        timestamp last_processed_at
    }
    
    COLUMN {
        bigint id PK
        bigint dataset_id FK
        string name
        string data_type
        boolean is_nullable
        string constraint_check
        integer position
        string description
    }
    
    DATASET_VERSION {
        bigint id PK
        bigint dataset_id FK
        string version_tag
        string operation_type
        string parameters_json
        long row_count
        timestamp created_at
    }
    
    OPERATION_LOG {
        bigint id PK
        bigint user_id FK
        string operation_type
        string entity_name
        string entity_id
        string parameters_json
        string result_status
        string error_message
        timestamp created_at
    }
```

### Data Structure Design

#### 1. **Core Data Structures (Go Structs)**

```go
// User Model
type User struct {
    ID        uint      `gorm:"primaryKey"`
    Username  string    `gorm:"uniqueIndex;not null"`
    Email     string    `gorm:"uniqueIndex;not null"`
    Password  string    `gorm:"not null"`
    IsActive  bool      `gorm:"default:true"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Dataset Model
type Dataset struct {
    ID             uint      `gorm:"primaryKey"`
    UserID         uint      `gorm:"index"`
    Name           string    `gorm:"not null"`
    Description    string
    OriginalFile   string    `gorm:"not null"`
    Format         string    `gorm:"not null"` // csv, json, xml, excel
    RowCount       int64
    FileSize       int64
    ProcessingMeta JSON      `gorm:"type:jsonb"`
    CreatedAt      time.Time
}

// DataRow - Generic row structure
type DataRow struct {
    RowIndex int                    `json:"row_index"`
    Values   map[string]interface{} `json:"values"`
    IsValid  bool                   `json:"is_valid"`
    Errors   []string               `json:"errors,omitempty"`
}

// FilterCondition
type FilterCondition struct {
    Column    string      `json:"column"`
    Operator  string      `json:"operator"` // >, <, ==, !=, contains, regex
    Value     interface{} `json:"value"`
    Logic     string      `json:"logic"` // AND, OR
}

// TransformOperation
type TransformOperation struct {
    Type       string      `json:"type"` // add, update, delete, rename
    Target     string      `json:"target"`
    Expression string      `json:"expression,omitempty"`
    Value      interface{} `json:"value,omitempty"`
}

// ValidationRule
type ValidationRule struct {
    Column     string   `json:"column"`
    Required   bool     `json:"required"`
    DataType   string   `json:"data_type"`
    MinLength  *int     `json:"min_length,omitempty"`
    MaxLength  *int     `json:"max_length,omitempty"`
    Pattern    string   `json:"pattern,omitempty"`
    EnumValues []string `json:"enum_values,omitempty"`
}
```

#### 2. **Process Design - Data Flow**
```mermaid
flowchart TD
    Start([Start Process]) --> Input[/Read Input File/]
    
    Input --> Parse{"Parse Format?"}
    Parse -->|CSV| CSVParser[CSV Parser<br/>encoding/csv]
    Parse -->|JSON| JSONParser[JSON Parser<br/>encoding/json]
    Parse -->|XML| XMLParser[XML Parser<br/>encoding/xml]
    Parse -->|Excel| ExcelParser[Excel Parser<br/>xuri/excelize]
    
    CSVParser --> Validate[Validate Data]
    JSONParser --> Validate
    XMLParser --> Validate
    ExcelParser --> Validate
    
    Validate --> Filter{"Filter Required?"}
    Filter -->|Yes| ApplyFilter[Apply Filter Conditions]
    Filter -->|No| Transform
    
    ApplyFilter --> Transform{"Transform Required?"}
    
    Transform -->|Yes| ApplyTransform[Apply Transformations]
    Transform -->|No| Export
    
    ApplyTransform --> Export
    
    Export --> Format{"Export Format?"}
    Format -->|CSV| CSVExport[CSV Export]
    Format -->|JSON| JSONExport[JSON Export]
    Format -->|XML| XMLExport[XML Export]
    Format -->|Excel| ExcelExport[Excel Export]
    Format -->|DB| DBExport[Database Export]
    
    CSVExport --> Result[/Save Result/]
    JSONExport --> Result
    XMLExport --> Result
    ExcelExport --> Result
    DBExport --> Result
    
    Result --> Log[Log Operation]
    Log --> End([End Process])
    
    style Start fill:#90EE90
    style End fill:#FFB6C1
    style Validate fill:#FFE4B5
```

#### 3. **Circuit Design - Error Handling Flow**
```mermaid
flowchart LR
    Operation[Operation Request] --> Try[Try Execute]
    
    Try --> Success{Success?}
    Success -->|Yes| ReturnOK[Return Result]
    Success -->|No| CatchError[Catch Error]
    
    CatchError --> LogError[Log Error Details]
    LogError --> CheckFatal{Is Fatal?}
    
    CheckFatal -->|Yes| FatalHandler[Fatal Error Handler
    - Rollback Transaction
    - Alert Admin
    - Return 500]
    
    CheckFatal -->|No| Recover[Recovery Handler
    - Retry Logic
    - Fallback Option
    - Return 400/422]
    
    FatalHandler --> EndError
    Recover --> EndError
    
    ReturnOK --> EndSuccess([Success])
    EndError --> EndFail([Error Response])
    
    classDef success fill:#90EE90,stroke:#006400
    classDef error fill:#FFB6C1,stroke:#8B0000
    classDef warn fill:#FFE4B5,stroke:#B8860B
    
    class ReturnOK,EndSuccess success
    class EndError,EndFail error
    class LogError,CheckFatal,Recover warn
```

### Index Strategy

| Table | Column(s) | Type | Purpose |
|-------|-----------|------|---------|
| users | username | UNIQUE | Fast user lookup |
| users | email | UNIQUE | Email-based auth |
| user_session | user_id, expires_at | INDEX | Session cleanup |
| dataset | user_id, created_at | INDEX | User dataset listing |
| operation_log | user_id, created_at | INDEX | Audit trail queries |
| column | dataset_id | INDEX | Column retrieval |

### Data Retention Policy
- Active datasets: Retained indefinitely
- Operation logs: Retained for 1 year
- User sessions: Retained for 30 days
- Deleted datasets: Soft delete (is_deleted flag), hard delete after 90 days

---

## 5.3 Input / Output and Interface Design

### 5.3.1 Input Specifications

#### CLI Input Format
```bash
# Filter Command
./datautil filter --input dataset.csv --where "age > 25 AND city = 'NYC'" --output filtered.csv

# Transform Command
./datautil transform --input data.csv --add "full_name=first_name+' '+last_name" --output transformed.csv

# Validate Command
./datautil validate --input file.csv --required "name,email" --types "string,email" --output report.json

# Export Command
./datautil export --input data.csv --to json --output data.json
```

#### API Input Format (JSON)
```json
{
  "input": "dataset.csv",
  "parameters": {
    "filter": {
      "conditions": [
        {
          "column": "age",
          "operator": ">",
          "value": 25
        }
      ]
    },
    "transform": [
      {
        "type": "add",
        "target": "full_name",
        "expression": "first_name + ' ' + last_name"
      }
    ]
  },
  "output_format": "csv",
  "output_path": "result.csv"
}
```

### 5.3.2 Output Specifications

#### CLI Output Format
```
✓ Operation completed successfully
  Input:  data.csv (1000 rows)
  Output: filtered.csv (850 rows)
  Time:   2.3s
  Rows processed: 1000
  Rows affected: 850
  Errors: 0
```

#### API Response Format
```json
{
  "status": "success",
  "message": "Operation completed successfully",
  "data": {
    "output_file": "result.csv",
    "rows_processed": 1000,
    "rows_affected": 850,
    "processing_time_ms": 2300,
    "errors": []
  },
  "metadata": {
    "request_id": "req_123456",
    "timestamp": "2026-04-27T12:00:00Z",
    "version": "1.0.0"
  }
}
```

### 5.3.3 Error Response Format
```json
{
  "status": "error",
  "message": "Validation failed",
  "error": {
    "code": "VALIDATION_ERROR",
    "details": [
      {
        "field": "email",
        "message": "Invalid email format",
        "row": 42
      }
    ]
  },
  "metadata": {
    "request_id": "req_123457",
    "timestamp": "2026-04-27T12:00:01Z"
  }
}
```

### 5.3.4 Interface Components

#### CLI Interface Components
1. **Flags & Options**
   - Global: `--config`, `--verbose`, `--output`
   - Auth: `--token`, `--username`, `--password`
   - Data: `--input`, `--where`, `--add`, `--validate`

2. **Interactive Prompts**
   - Password input (hidden)
   - Confirmation for destructive operations
   - Progress bars for long operations

3. **Help System**
   - Auto-generated from cobra commands
   - Examples for each command
   - Flag descriptions

#### HTTP API Interface
- **Base URL**: `http://localhost:8080/api`
- **Content-Type**: `application/json`
- **Authentication**: `Bearer <token>` header
- **Swagger UI**: `http://localhost:8080/swagger`

### 5.3.5 State Transition Diagram
```mermaid
stateDiagram-v2
    [*] --> Idle: System Started
    
    Idle --> Authenticating: Auth Request
    Authenticating --> Authenticated: Valid Credentials
    Authenticating --> Idle: Invalid Credentials
    
    Authenticated --> Processing: Data Operation Request
    Authenticated --> Idle: Logout
    
    Processing --> Validating: Input Received
    Validating --> Filtering: Valid Input
    Validating --> Error: Invalid Input
    
    Filtering --> Transforming: Filter Complete
    Filtering --> Exporting: No Transform
    
    Transforming --> Exporting: Transform Complete
    
    Exporting --> Saving: Export Complete
    Saving --> Complete: Save Success
    Saving --> Error: Save Failed
    
    Error --> Idle: Error Handled
    Complete --> Idle: Operation Done
    
    Idle --> [*]: System Shutdown
    
    note right of Processing: Operation States
    note right of Validating: Data Validation
    note right of Filtering: Row Filtering
    note right of Transforming: Data Transform
    note right of Exporting: Format Export
    note right of Saving: Result Persistence
```

---

## 5.3.2 Samples of Forms, Reports and Interface

### CLI Interface Samples

#### 1. Login Form (Interactive)
```
$ ./datautil login

  DataUtil CLI - Authentication
═══════════════════════════════════

  Email: admin@example.com
  Password: ********
  
  [Login] [Cancel]

→ Logging in...
✓ Authentication successful
  Token: eyJhbGc... (expires in 24h)
  User:  Admin User
```

#### 2. Data Filter Form
```
$ ./datautil filter --help

Filter dataset rows based on conditions

Usage:
  datautil filter [flags]

Flags:
  -i, --input string       Input file path (required)
  -w, --where string       Filter condition (e.g., "age > 25")
  -o, --output string      Output file path
  -f, --format string      Output format (csv, json, xml, excel)
  -l, --limit int          Maximum rows to return
  -h, --help               Help for filter

Examples:
  # Filter by age
  datautil filter --input data.csv --where "age > 25"
  
  # Filter by multiple conditions
  datautil filter --input data.csv --where "age > 25 AND city = 'NYC'"
  
  # Filter with regex
  datautil filter --input data.csv --where "email CONTAINS '@gmail.com'"
```

#### 3. Operation Report Sample
```
══════════════════════════════════════════════════════════
              DATAUTIL - OPERATION REPORT
══════════════════════════════════════════════════════════

Operation: FILTER
Date: 2026-04-27 12:00:00
User: admin@example.com

INPUT FILE:
  Path: /data/input.csv
  Size: 2.5 MB
  Rows: 10,000
  Columns: 12

FILTER CONDITIONS:
  ├─ age > 25
  ├─ status = 'active'
  └─ country IN ('USA', 'CANADA')

OUTPUT:
  Path: /data/output.csv
  Size: 1.8 MB
  Rows: 6,234 (62.34%)
  Columns: 12

PROCESSING:
  Duration: 1.45s
  Rows/sec: 6,896
  Memory: 45.2 MB
  Threads: 4

══════════════════════════════════════════════════════════
```

### HTTP API Interface Samples

#### 1. API Endpoint: Filter Data
```http
POST /api/data/filter
Authorization: Bearer eyJhbGc...
Content-Type: application/json

{
  "input": "dataset.csv",
  "filter": {
    "conditions": [
      {
        "column": "age",
        "operator": ">",
        "value": 25
      },
      {
        "column": "status",
        "operator": "=",
        "value": "active"
      }
    ],
    "logic": "AND"
  },
  "output_format": "csv"
}
```

#### 2. API Response
```json
{
  "status": "success",
  "message": "Filter operation completed",
  "data": {
    "output_file": "filtered_dataset_20260427.csv",
    "download_url": "/api/files/download/filtered_dataset_20260427.csv",
    "rows_processed": 10000,
    "rows_returned": 6234,
    "processing_time_ms": 1450
  },
  "metadata": {
    "request_id": "req_abc123",
    "timestamp": "2026-04-27T12:00:00Z",
    "version": "1.0.0"
  }
}
```

### Web Interface (Swagger UI)
```
Swagger Documentation
├─ /api/health
│  └─ GET Health check endpoint
├─ /api/auth/register
│  └─ POST Register new user
├─ /api/auth/login
│  └─ POST Authenticate and get token
├─ /api/data/filter
│  └─ POST Filter dataset rows
├─ /api/data/transform
│  └─ POST Transform dataset
├─ /api/data/validate
│  └─ POST Validate dataset
├─ /api/data/export
│  └─ POST Export dataset
└─ /api/db/query
   └─ POST Execute SQL query
```

---

## 5.3.3 Access Control / Mechanism / Security

### Authentication Flow
```mermaid
sequenceDiagram
    participant Client
    participant API
    participant Auth
    participant DB
    
    Client->>API: POST /api/auth/login
    API->>Auth: Validate credentials
    Auth->>DB: Query user
    DB-->>Auth: User data
    Auth->>Auth: Verify password (bcrypt)
    Auth->>Auth: Generate JWT token
    Auth-->>API: Return token
    API-->>Client: 200 OK {token}
    
    Client->>API: Request (with Bearer token)
    API->>Auth: Verify token
    Auth->>Auth: Check expiry & signature
    Auth-->>API: Token valid
    API->>API: Process request
    API-->>Client: 200 OK {data}
```

### Authorization Matrix

| Role | Read Data | Write Data | Filter | Transform | Validate | Export | Manage Users | System Config |
|------|----------|-----------|--------|-----------|----------|--------|--------------|---------------|
| Guest | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ |
| User | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✗ | ✗ |
| Admin | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | Limited |
| SuperAdmin | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |

### Security Implementation

#### 1. **JWT Token Structure**
```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "sub": "user-id",
    "username": "admin",
    "role": "admin",
    "permissions": ["read", "write", "delete"],
    "iat": 1651036800,
    "exp": 1651123200,
    "jti": "unique-token-id"
  },
  "signature": "HMACSHA256(base64UrlEncode(header)+'.'+base64UrlEncode(payload), secret)"
}
```

#### 2. **Password Security**
- Algorithm: bcrypt (cost factor: 12)
- Minimum length: 8 characters
- Requirements: uppercase, lowercase, number, special character
- Storage: hashed only (never plaintext)

#### 3. **API Rate Limiting**
```go
type RateLimiter struct {
    RequestsPerMinute int
    BurstSize        int
    Enabled          bool
}

// Default limits
{
    "guest":  {"rpm": 60, "burst": 10},
    "user":   {"rpm": 300, "burst": 50},
    "admin":  {"rpm": 1000, "burst": 100},
}
```

#### 4. **Input Validation & Sanitization**
```mermaid
flowchart TD
    Input[User Input] --> Sanitize[Remove malicious chars]
    Sanitize --> Validate[Type/Format Check]
    Validate --> Length{Length OK?}
    Length -->|Yes| Range{Value in Range?}
    Length -->|No| Reject[Reject Request]
    Range -->|Yes| SQLSafe[SQL Injection Check]
    Range -->|No| Reject
    SQLSafe --> XSSCheck[XSS Check]
    XSSCheck --> Safe[Process Input]
```

### Security Features

#### **Data Encryption**
- **At Rest**: AES-256 encryption for sensitive fields
- **In Transit**: TLS 1.3 for all API communications
- **Key Management**: Environment variables + KMS integration

#### **Audit Logging**
```go
type AuditLog struct {
    Timestamp   time.Time `json:"timestamp"`
    UserID      uint      `json:"user_id"`
    Action      string    `json:"action"`
    Resource    string    `json:"resource"`
    Status      string    `json:"status"`
    IPAddress   string    `json:"ip_address"`
    UserAgent   string    `json:"user_agent"`
    RequestID   string    `json:"request_id"`
}
```

#### **Session Management**
- Token expiry: 24 hours (configurable)
- Refresh token: 7 days
- Concurrent sessions: Maximum 5
- Auto-logout: 30 minutes inactivity
- Token revocation: Immediate on logout

#### **Security Headers**
```
HTTP/1.1 200 OK
Content-Security-Policy: default-src 'self'
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
Strict-Transport-Security: max-age=31536000; includeSubDomains
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
```

### Access Control Implementation

#### **Role-Based Access Control (RBAC)**
```go
type Permission string

const (
    PermissionReadData   Permission = "data:read"
    PermissionWriteData  Permission = "data:write"
    PermissionDeleteData Permission = "data:delete"
    PermissionManageUser Permission = "user:manage"
    PermissionSystemAdmin Permission = "system:admin"
)

type Role struct {
    Name        string        `json:"name"`
    Permissions []Permission  `json:"permissions"`
}

// Role definitions
var Roles = map[string]Role{
    "guest": {
        Name:        "Guest",
        Permissions: []Permission{PermissionReadData},
    },
    "user": {
        Name:        "User",
        Permissions: []Permission{
            PermissionReadData,
            PermissionWriteData,
        },
    },
    "admin": {
        Name:        "Admin",
        Permissions: []Permission{
            PermissionReadData,
            PermissionWriteData,
            PermissionDeleteData,
            PermissionManageUser,
        },
    },
    "superadmin": {
        Name:        "Super Admin",
        Permissions: []Permission{
            PermissionReadData,
            PermissionWriteData,
            PermissionDeleteData,
            PermissionManageUser,
            PermissionSystemAdmin,
        },
    },
}
```

#### **Middleware for Authorization**
```go
func RequirePermission(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        user := c.MustGet("user").(*User)
        
        if !user.HasPermission(permission) {
            c.JSON(403, gin.H{
                "error": "insufficient_permissions",
                "message": "You don't have permission to perform this action",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### Security Monitoring

#### **Real-time Alerts**
- Failed login attempts (>5 in 5 minutes)
- Unusual access patterns (geographic anomalies)
- Privilege escalation attempts
- Data exfiltration patterns (bulk exports)
- Brute force detection

#### **Compliance**
- GDPR compliant data handling
- Right to erasure implementation
- Data portability features
- Consent management
- Data retention policies

### Vulnerability Protection

| Threat | Mitigation |
|--------|-----------|
| SQL Injection | Parameterized queries, ORM |
| XSS Attacks | Input sanitization, output encoding |
| CSRF | SameSite cookies, anti-CSRF tokens |
| Brute Force | Rate limiting, account lockout |
| DDoS | Rate limiting, CDN, WAF |
| MITM | TLS 1.3, HSTS enforcement |
| IDOR | UUID resource IDs, access checks |
| Token Theft | Short expiry, refresh tokens, HttpOnly cookies |