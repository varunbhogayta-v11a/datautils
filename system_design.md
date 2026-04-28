### 5.4.3 Data Format Support

#### CSV Format

**Structure:**
```
header1,header2,header3
value1,value2,value3
value4,value5,value6
```

**Reader:** `encoding/csv` with default comma delimiter  
**Writer:** `encoding/csv` with comma delimiter

---

#### JSON Format

**Structure (Array of Objects):**
```json
[
  {
    "name": "John",
    "age": 25,
    "city": "NYC"
  },
  {
    "name": "Jane",
    "age": 30,
    "city": "LA"
  }
]
```

**Reader:** `encoding/json` → `[]map[string]interface{}`  
**Headers:** Extracted from all keys across objects  
**Writer:** `json.Encoder` with indentation

---

#### XML Format

**Structure:**
```xml
<data>
  <record>
    <name>John</name>
    <age>25</age>
  </record>
  <record>
    <name>Jane</name>
    <age>30</age>
  </record>
</data>
```

**Reader:** `encoding/xml` with custom `xmlRecord` struct  
**Headers:** Extracted from field names in first record  
**Writer:** Custom XML encoder with nested structure

---

#### Excel Format (.xlsx)

**Structure:** Read first sheet only  
**Reader:** `github.com/xuri/excelize/v2`  
- Opens file  
- Gets first sheet name  
- Reads all rows  
- First row = headers, rest = data  
**Writer:** NOT implemented (read-only)

---

### 5.4.4 Configuration Interface

#### Configuration Sources (Priority Order)

1. **Explicit flag:** `--config=/path/to/config.yaml`
2. **Current directory:** `./config.yaml`
3. **Configs directory:** `./configs/config.yaml`
4. **Home directory:** `$HOME/.datautil/config.yaml`
5. **Environment variables:** Override all above

#### Environment Variables

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `DB_HOST` | localhost | No | PostgreSQL host |
| `DB_PORT` | 5432 | No | PostgreSQL port |
| `DB_USER` | postgres | No | Database username |
| `DB_PASSWORD` | postgres | No | Database password |
| `DB_NAME` | datautil | No | Database name |
| `DB_DRIVER` | sqlite3 | No | Driver: postgres/sqlite/mysql |
| `JWT_SECRET` | (none) | **Yes for prod** | JWT signing secret |
| `JWT_EXPIRY` | 24h | No | Token expiration duration |
| `PORT` | 8080 | No | API server port |
| `HOST` | 0.0.0.0 | No | API server host |

**Note:** In production, `JWT_SECRET` must be set. Default is only for development.

---

## 5.5 STATE TRANSITION DIAGRAMS FOR AUTHENTICATION FLOW

### 5.5.1 User Registration State Diagram

```
                    ┌──────────────┐
                    │  Start:      │
                    │  User wants  │
                    │  to register │
                    └──────┬───────┘
                           │
                           ▼
                    ┌──────────────┐
                    │ Submit       │
                    │ registration │
                    │ form with    │
                    │ username,    │
                    │ email,       │
                    │ password     │
                    └──────┬───────┘
                           │
                           ▼
                    ┌──────────────┐
                    │ Validate     │
                    │ input:       │
                    │ - Non-empty  │
                    │ - Email      │
                    │   format     │
                    │ - Username   │
                    │   available  │
                    └──────┬───────┘
                           │
               ┌───────────┴───────────┐
               │                       │
        VALID │                       │ INVALID
               │                       │
               ▼                       ▼
        ┌──────────────┐        ┌──────────────┐
        │ Check if     │        │ Return error │
        │ user exists  │        │ message:     │
        │ by email/    │        │ "User already│
        │ username     │        │  exists"     │
        └──────┬───────┘        └──────────────┘
               │
               ▼
        ┌──────────────┐
        │ Hash         │
        │ password     │
        │ with bcrypt  │
        │ (cost=14)    │
        └──────┬───────┘
               │
               ▼
        ┌──────────────┐
        │ Create user  │
        │ record in    │
        │ database:    │
        │ - username   │
        │ - email      │
        │ - password   │
        │   (hashed)   │
        │ - role=user  │
        │ - active=true│
        └──────┬───────┘
               │
               ▼
        ┌──────────────┐
        │ Return       │
        │ success:     │
        │ - user ID    │
        │ - username   │
        │ - email      │
        │ - role       │
        └──────┬───────┘
               │
               ▼
        ┌──────────────┐
        │  End: User   │
        │  registered  │
        │  successfully│
        └──────────────┘
```

### 5.5.2 User Login State Diagram

```
                    ┌──────────────┐
                    │  Start:      │
                    │  User wants  │
                    │  to login    │
                    └──────┬───────┘
                           │
                           ▼
                    ┌──────────────┐
                    │ Submit       │
                    │ login form   │
                    │ with email   │
                    │ and password │
                    └──────┬───────┘
                           │
                           ▼
                    ┌──────────────┐
                    │ Validate     │
                    │ credentials: │
                    │ - Email not  │
                    │   empty      │
                    │ - Password   │
                    │   not empty  │
                    └──────┬───────┘
                           │
               ┌───────────┴───────────┐
               │                       │
        VALID │                       │ INVALID
               │                       │
               ▼                       ▼
        ┌──────────────┐        ┌──────────────┐
        │ Look up user │        │ Return error │
        │ by email     │        │ "Invalid     │
        │ in database  │        │  credentials"│
        └──────┬───────┘        └──────────────┘
               │
               ▼
        ┌──────────────┐
        │ User found?  │
        └──────┬───────┘
               │
       ┌───────┴───────┐
       │               │
   YES │               │ NO
       │               │
       ▼               ▼
┌──────────────┐  ┌──────────────┐
│ Check if     │  │ Return error │
│ account is   │  │ "User not    │
│ active       │  │  found"      │
└──────┬───────┘  └──────────────┘
       │
       ▼
┌──────────────┐
│ Compare      │
│ password:    │
│ bcrypt.      │
│ CompareHash  │
│ AndPassword  │
└──────┬───────┘
       │
       ├─────────────┐
       │             │
   MATCH           NO MATCH
       │             │
       ▼             ▼
┌──────────────┐  ┌──────────────┐
│ Generate JWT │  │ Return error │
│ token:       │  │ "Invalid     │
│ - user_id    │  │  credentials"│
│ - username   │  │              │
│ - role       │  │              │
│ - exp = now  │  │              │
│   + 24h      │  │              │
└──────┬───────┘  └──────────────┘
       │
       ▼
┌──────────────┐
│ Return:      │
│ - User info  │
│ - JWT token  │
└──────┬───────┘
       │
       ▼
    ┌──────────────┐
    │  End: User   │
    │  logged in   │
    │  with token  │
    └──────────────┘
```

### 5.5.3 JWT Token Validation State Diagram

```
                    ┌──────────────┐
                    │  Start:      │
                    │  API request │
                    │  with token  │
                    └──────┬───────┘
                           │
                           ▼
                    ┌──────────────┐
                    │ Extract      │
                    │ token from   │
                    │ Authorization│
                    │ header or    │
                    │ query param  │
                    └──────┬───────┘
                           │
                           ▼
                    ┌──────────────┐
                    │ Token empty? │
                    └──────┬───────┘
                           │
               ┌───────────┴───────────┐
               │                       │
           YES │                       │ NO
               │                       │
               ▼                       ▼
        ┌──────────────┐        ┌──────────────┐
        │ Return       │        │ Parse token  │
        │ error:       │        │ with         │
        │ "Auth        │        │ jwt.Parse    │
        │  required"   │        │ WithClaims   │
        └──────────────┘        └──────┬───────┘
                                     │
                                     ▼
                              ┌──────────────┐
                              │ Validate     │
                              │ signature:   │
                              │ - Check alg  │
                              │   is HS256   │
                              │ - Verify     │
                              │   signature  │
                              │   with secret│
                              └──────┬───────┘
                                     │
                           ┌─────────┴─────────┐
                           │                   │
                     VALID │                   │ INVALID
                           │                   │
                           ▼                   ▼
                    ┌──────────────┐    ┌──────────────┐
                    │ Check claims │    │ Return error │
                    │ - NotBefore  │    │ "Invalid or  │
                    │ - ExpiresAt  │    │  expired     │
                    │ - IssuedAt   │    │  token"      │
                    └──────┬───────┘    └──────────────┘
                           │
                           ▼
                    ┌──────────────┐
                    │ Extract      │
                    │ custom claims│
                    │ - user_id    │
                    │ - username   │
                    │ - role       │
                    └──────┬───────┘
                           │
                           ▼
                    ┌──────────────┐
                    │ Return       │
                    │ *Claims      │
                    │ struct with  │
                    │ user context │
                    └──────┬───────┘
                           │
                           ▼
                    ┌──────────────┐
                    │  Token       │
                    │  accepted    │
                    │  - Continue  │
                    │    to handler│
                    └──────────────┘
```