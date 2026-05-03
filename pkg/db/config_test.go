package db

import (
	"os"
	"testing"
)

func TestGetConfig(t *testing.T) {
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

	cfg := GetConfig()
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

func TestGetConfig_Defaults(t *testing.T) {
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("DB_DRIVER")
	os.Unsetenv("DB_SSLMODE")

	cfg := GetConfig()
	if cfg.Host != "localhost" {
		t.Errorf("Host = %q, want %q", cfg.Host, "localhost")
	}
	if cfg.Port != "5432" {
		t.Errorf("Port = %q, want %q", cfg.Port, "5432")
	}
	if cfg.User != "postgres" {
		t.Errorf("User = %q, want %q", cfg.User, "postgres")
	}
	if cfg.Driver != "sqlite3" {
		t.Errorf("Driver = %q, want %q", cfg.Driver, "sqlite3")
	}
}

func TestConfig_DriverName(t *testing.T) {
	tests := []struct {
		driver string
		want   string
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
			cfg := &Config{Driver: tt.driver}
			if got := cfg.DriverName(); got != tt.want {
				t.Errorf("DriverName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConfig_ConnString(t *testing.T) {
	t.Run("postgres", func(t *testing.T) {
		cfg := &Config{
			Driver:   "postgres",
			Host:     "localhost",
			Port:     "5432",
			User:     "user",
			Password: "pass",
			DBName:   "testdb",
			SSLMode:  "disable",
		}
		connStr := cfg.ConnString()
		if connStr == "" {
			t.Error("expected non-empty connection string")
		}
	})

	t.Run("mysql", func(t *testing.T) {
		cfg := &Config{
			Driver:   "mysql",
			Host:     "localhost",
			Port:     "3306",
			User:     "user",
			Password: "pass",
			DBName:   "testdb",
		}
		connStr := cfg.ConnString()
		if connStr == "" {
			t.Error("expected non-empty connection string")
		}
	})

	t.Run("sqlite", func(t *testing.T) {
		cfg := &Config{
			Driver: "sqlite",
			DBName: "test.db",
		}
		connStr := cfg.ConnString()
		if connStr != "test.db" && connStr != "test.db.db" {
			t.Errorf("ConnString() = %q, want %q", connStr, "test.db")
		}
	})
}

func TestGetEnv(t *testing.T) {
	key := "TEST_ENV_KEY"
	value := "test_value"

	os.Setenv(key, value)
	defer os.Unsetenv(key)

	if got := getEnv(key, "default"); got != value {
		t.Errorf("getEnv(%q) = %q, want %q", key, got, value)
	}

	if got := getEnv("NON_EXISTENT_"+key, "default"); got != "default" {
		t.Errorf("getEnv() = %q, want %q", got, "default")
	}
}

func TestGetDB(t *testing.T) {
	db := GetDB()
	_ = db
}

func TestGetDriver(t *testing.T) {
	driver := GetDriver()
	if driver == "" {
		t.Error("expected non-empty driver")
	}
}
