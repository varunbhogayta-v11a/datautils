package models

import (
	"time"

	"github.com/doug-martin/goqu/v9"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
	RoleGuest Role = "guest"
)

type User struct {
	ID        uint      `db:"id" goqu:"skipinsert"`
	Username  string    `db:"username"`
	Email     string    `db:"email"`
	Password  string    `db:"password"`
	Role      string    `db:"role"`
	Active    bool      `db:"active"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type UserTable struct{}

var UserTableObj = &UserTable{}

func (*UserTable) Insert() *goqu.InsertDataset {
	return goqu.Insert("users")
}

func (*UserTable) Update() *goqu.UpdateDataset {
	return goqu.Update("users")
}

func (*UserTable) Delete() *goqu.DeleteDataset {
	return goqu.Delete("users")
}

type OperationLog struct {
	ID         uint      `db:"id" goqu:"skipinsert"`
	UserID     uint      `db:"user_id"`
	Operation  string    `db:"operation"`
	InputFile  string    `db:"input_file"`
	OutputFile string    `db:"output_file"`
	Details    string    `db:"details"`
	CreatedAt  time.Time `db:"created_at"`
}

type OperationLogTable struct{}

var OperationLogTableObj = &OperationLogTable{}

func (*OperationLogTable) Insert() *goqu.InsertDataset {
	return goqu.Insert("operation_logs")
}

func GetTableName(driver string) map[string]string {
	switch driver {
	case "postgres", "pgx":
		return map[string]string{
			"users":          "CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, username VARCHAR(50) UNIQUE NOT NULL, email VARCHAR(100) UNIQUE NOT NULL, password VARCHAR(255) NOT NULL, role VARCHAR(20) DEFAULT 'user', active BOOLEAN DEFAULT true, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)",
			"operation_logs": "CREATE TABLE IF NOT EXISTS operation_logs (id SERIAL PRIMARY KEY, user_id INTEGER NOT NULL, operation VARCHAR(50) NOT NULL, input_file VARCHAR(255), output_file VARCHAR(255), details TEXT, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, FOREIGN KEY (user_id) REFERENCES users(id))",
		}
	case "mysql":
		return map[string]string{
			"users":          "CREATE TABLE IF NOT EXISTS users (id INT AUTO_INCREMENT PRIMARY KEY, username VARCHAR(50) UNIQUE NOT NULL, email VARCHAR(100) UNIQUE NOT NULL, password VARCHAR(255) NOT NULL, role VARCHAR(20) DEFAULT 'user', active BOOLEAN DEFAULT true, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP)",
			"operation_logs": "CREATE TABLE IF NOT EXISTS operation_logs (id INT AUTO_INCREMENT PRIMARY KEY, user_id INT NOT NULL, operation VARCHAR(50) NOT NULL, input_file VARCHAR(255), output_file VARCHAR(255), details TEXT, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, FOREIGN KEY (user_id) REFERENCES users(id))",
		}
	case "sqlite", "sqlite3":
		return map[string]string{
			"users":          "CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, username VARCHAR(50) UNIQUE NOT NULL, email VARCHAR(100) UNIQUE NOT NULL, password VARCHAR(255) NOT NULL, role VARCHAR(20) DEFAULT 'user', active INTEGER DEFAULT 1, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP)",
			"operation_logs": "CREATE TABLE IF NOT EXISTS operation_logs (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER NOT NULL, operation VARCHAR(50) NOT NULL, input_file VARCHAR(255), output_file VARCHAR(255), details TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, FOREIGN KEY (user_id) REFERENCES users(id))",
		}
	default:
		return map[string]string{}
	}
}

func (u *User) HasPermission(action string) bool {
	switch u.Role {
	case "admin":
		return true
	case "user":
		return action != "delete" && action != "admin"
	case "guest":
		return action == "read"
	default:
		return false
	}
}
