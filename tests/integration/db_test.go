//go:build integration
// +build integration

package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/varunbhogayta-v11a/datautils/pkg/db"
)

func TestIntegration_ConnectSQLite(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	os.Setenv("DB_DRIVER", "sqlite3")
	os.Setenv("DB_NAME", dbPath)
	defer os.Unsetenv("DB_DRIVER")
	defer os.Unsetenv("DB_NAME")

	cfg := db.GetConfig()
	if err := db.Connect(cfg); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer db.Close()

	if db.GetDB() == nil {
		t.Error("expected DB to be non-nil")
	}
}

func TestConfig_GetConfig(t *testing.T) {
	os.Setenv("DB_HOST", "myhost")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "myuser")
	os.Setenv("DB_PASSWORD", "mypass")
	os.Setenv("DB_NAME", "mydb")
	os.Setenv("DB_DRIVER", "sqlite3")
	os.Setenv("DB_SSLMODE", "require")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_DRIVER")
		os.Unsetenv("DB_SSLMODE")
	}()

	cfg := db.GetConfig()
	if cfg.Host != "myhost" {
		t.Errorf("Host = %q, want %q", cfg.Host, "myhost")
	}
	if cfg.Port != "5433" {
		t.Errorf("Port = %q, want %q", cfg.Port, "5433")
	}
	if cfg.User != "myuser" {
		t.Errorf("User = %q, want %q", cfg.User, "myuser")
	}
	if cfg.Driver != "sqlite3" {
		t.Errorf("Driver = %q, want %q", cfg.Driver, "sqlite3")
	}
}

func TestConfig_DriverName(t *testing.T) {
	tests := []struct {
		driver  string
		want    string
	}{
		{"postgres", "pgx"},
		{"pgx", "pgx"},
		{"mysql", "mysql"},
		{"sqlite", "sqlite3"},
		{"sqlite3", "sqlite3"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			cfg := &db.Config{Driver: tt.driver}
			if got := cfg.DriverName(); got != tt.want {
				t.Errorf("DriverName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConfig_ConnString(t *testing.T) {
	t.Run("postgres", func(t *testing.T) {
		cfg := &db.Config{
			Driver:   "postgres",
			Host:    "localhost",
			Port:    "5432",
			User:    "user",
			Password: "pass",
			DBName:  "testdb",
			SSLMode: "disable",
		}
		connStr := cfg.ConnString()
		if connStr == "" {
			t.Error("expected non-empty connection string")
		}
	})

	t.Run("mysql", func(t *testing.T) {
		cfg := &db.Config{
			Driver:   "mysql",
			Host:    "localhost",
			Port:    "3306",
			User:    "user",
			Password: "pass",
			DBName:  "testdb",
		}
		connStr := cfg.ConnString()
		if connStr == "" {
			t.Error("expected non-empty connection string")
		}
	})

	t.Run("sqlite", func(t *testing.T) {
		cfg := &db.Config{
			Driver: "sqlite",
			DBName: "test.db",
		}
		connStr := cfg.ConnString()
		if connStr != "test.db" && connStr != "test.db.db" {
			t.Errorf("ConnString() = %q, want %q", connStr, "test.db")
		}
	})
}

func TestDB_Close(t *testing.T) {
	db.Close()
}

func TestDB_GetDriver(t *testing.T) {
	driver := db.GetDriver()
	if driver == "" {
		t.Error("expected non-empty driver")
	}
}

func TestIntegration_CreateTables(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	os.Setenv("DB_DRIVER", "sqlite3")
	os.Setenv("DB_NAME", dbPath)
	defer os.Unsetenv("DB_DRIVER")
	defer os.Unsetenv("DB_NAME")

	cfg := db.GetConfig()
	if err := db.Connect(cfg); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer db.Close()

	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username VARCHAR(50) UNIQUE NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		role VARCHAR(20) DEFAULT 'user',
		active INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := db.GetDB().Exec(createUsersTable)
	if err != nil {
		t.Fatalf("failed to create users table: %v", err)
	}
}

func TestIntegration_UserCRUD(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	os.Setenv("DB_DRIVER", "sqlite3")
	os.Setenv("DB_NAME", dbPath)
	defer os.Unsetenv("DB_DRIVER")
	defer os.Unsetenv("DB_NAME")

	cfg := db.GetConfig()
	if err := db.Connect(cfg); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer db.Close()

	createTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username VARCHAR(50) UNIQUE NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		role VARCHAR(20) DEFAULT 'user',
		active INTEGER DEFAULT 1
	);
	`
	_, _ = db.GetDB().Exec(createTable)

	_, err := db.GetDB().Exec("INSERT INTO users (username, email, password, role) VALUES (?, ?, ?, ?)",
		"testuser", "test@example.com", "hashpass", "user")
	if err != nil {
		t.Fatalf("Insert error: %v", err)
	}

	var id int
	var username, email, role string
	err = db.GetDB().QueryRow("SELECT id, username, email, role FROM users WHERE username = ?", "testuser").
		Scan(&id, &username, &email, &role)
	if err != nil {
		t.Fatalf("QueryRow error: %v", err)
	}

	if username != "testuser" {
		t.Errorf("username = %q, want %q", username, "testuser")
	}

	_, err = db.GetDB().Exec("UPDATE users SET role = ? WHERE username = ?", "admin", "testuser")
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}

	err = db.GetDB().QueryRow("SELECT role FROM users WHERE username = ?", "testuser").Scan(&role)
	if err != nil {
		t.Fatalf("QueryRow error after update: %v", err)
	}

	if role != "admin" {
		t.Errorf("role = %q, want %q", role, "admin")
	}

	_, err = db.GetDB().Exec("DELETE FROM users WHERE username = ?", "testuser")
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}
}

func TestIntegration_OperationLogsCRUD(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	os.Setenv("DB_DRIVER", "sqlite3")
	os.Setenv("DB_NAME", dbPath)
	defer os.Unsetenv("DB_DRIVER")
	defer os.Unsetenv("DB_NAME")

	cfg := db.GetConfig()
	if err := db.Connect(cfg); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer db.Close()

	setup := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username VARCHAR(50) UNIQUE NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		role VARCHAR(20) DEFAULT 'user'
	);
	CREATE TABLE IF NOT EXISTS operation_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		operation VARCHAR(50) NOT NULL,
		input_file VARCHAR(255),
		output_file VARCHAR(255),
		details TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, _ = db.GetDB().Exec(setup)

	_, _ = db.GetDB().Exec("INSERT INTO users (username, email, password, role) VALUES (?, ?, ?, ?)",
		"testuser", "test@example.com", "hash", "user")

	var userID int
	_ = db.GetDB().QueryRow("SELECT id FROM users WHERE username = ?", "testuser").Scan(&userID)

	_, err := db.GetDB().Exec(`INSERT INTO operation_logs (user_id, operation, input_file, output_file, details)
		VALUES (?, ?, ?, ?, ?)`,
		userID, "filter", "input.csv", "output.csv", "Filtered 10 rows")
	if err != nil {
		t.Fatalf("Insert operation log error: %v", err)
	}

	var count int
	err = db.GetDB().QueryRow("SELECT COUNT(*) FROM operation_logs WHERE user_id = ?", userID).Scan(&count)
	if err != nil {
		t.Fatalf("Query operation logs error: %v", err)
	}

	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}