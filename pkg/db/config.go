package db

import (
	"fmt"
	"os"
	"time"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	Driver   string
	MaxConns int32
}

func GetConfig() *Config {
	return &Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "datautil"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
		Driver:   getEnv("DB_DRIVER", "sqlite3"),
		MaxConns: 25,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) DriverName() string {
	switch c.Driver {
	case "postgres", "pgx":
		return "pgx"
	case "mysql":
		return "mysql"
	case "sqlite", "sqlite3":
		return "sqlite3"
	default:
		return c.Driver
	}
}

func (c *Config) ConnString() string {
	switch c.Driver {
	case "postgres":
		return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
			c.User, c.Password, c.Host, c.Port, c.DBName)
	case "sqlite":
		return c.DBName + ".db"
	default:
		return c.DBName + ".db"
	}
}

func Connect(cfg *Config) error {
	var err error

	DB, err = sql.Open(cfg.DriverName(), cfg.ConnString())
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	DB.SetMaxOpenConns(int(cfg.MaxConns))
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(time.Hour)

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

func ConnectDefault() error {
	return Connect(GetConfig())
}

func CreateDatabase(cfg *Config) error {
	switch cfg.Driver {
	case "postgres":
		tempDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=%s",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.SSLMode)
		tempDB, err := sql.Open("pgx", tempDSN)
		if err != nil {
			return err
		}
		defer tempDB.Close()
		_, _ = tempDB.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.DBName))

	case "mysql":
		tempDB, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/?parseTime=true",
			cfg.User, cfg.Password, cfg.Host, cfg.Port))
		if err != nil {
			return err
		}
		defer tempDB.Close()
		_, _ = tempDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", cfg.DBName))

	case "sqlite":
		return nil
	}

	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}

func GetDB() *sql.DB    { return DB }
func GetDriver() string { return GetConfig().Driver }
