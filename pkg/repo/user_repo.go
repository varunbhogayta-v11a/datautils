package repo

import (
	"github.com/varunbhogayta-v11a/datautils/pkg/db"
	"github.com/varunbhogayta-v11a/datautils/pkg/models"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) Create(user *models.User) (*models.User, error) {
	query := `INSERT INTO users (username, email, password, role, active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
	result, err := db.DB.Exec(query, user.Username, user.Email, user.Password, user.Role, user.Active, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	user.ID = uint(id)
	return user, nil
}

func (r *UserRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	query := `SELECT id, username, email, password, role, active, created_at, updated_at FROM users WHERE id = ?`
	row := db.DB.QueryRow(query, id)
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Role, &user.Active, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	query := `SELECT id, username, email, password, role, active, created_at, updated_at FROM users WHERE email = ?`
	row := db.DB.QueryRow(query, email)
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Role, &user.Active, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	query := `SELECT id, username, email, password, role, active, created_at, updated_at FROM users WHERE username = ?`
	row := db.DB.QueryRow(query, username)
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Role, &user.Active, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(user *models.User) (*models.User, error) {
	query := `UPDATE users SET username = ?, email = ?, password = ?, role = ?, active = ?, updated_at = ? WHERE id = ?`
	_, err := db.DB.Exec(query, user.Username, user.Email, user.Password, user.Role, user.Active, user.UpdatedAt, user.ID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) Delete(id uint) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := db.DB.Exec(query, id)
	return err
}

func (r *UserRepository) List() ([]models.User, error) {
	query := `SELECT id, username, email, password, role, active, created_at, updated_at FROM users`
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Role, &user.Active, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

var _ models.UserRepository = (*UserRepository)(nil)

type OperationLogRepository struct{}

func NewOperationLogRepository() *OperationLogRepository {
	return &OperationLogRepository{}
}

func (r *OperationLogRepository) Create(log *models.OperationLog) (*models.OperationLog, error) {
	query := `INSERT INTO operation_logs (user_id, operation, input_file, output_file, details, created_at) VALUES (?, ?, ?, ?, ?, ?)`
	result, err := db.DB.Exec(query, log.UserID, log.Operation, log.InputFile, log.OutputFile, log.Details, log.CreatedAt)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	log.ID = uint(id)
	return log, nil
}

func (r *OperationLogRepository) GetByUserID(userID uint) ([]models.OperationLog, error) {
	query := `SELECT id, user_id, operation, input_file, output_file, details, created_at FROM operation_logs WHERE user_id = ? ORDER BY created_at DESC`
	rows, err := db.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.OperationLog
	for rows.Next() {
		var log models.OperationLog
		err := rows.Scan(&log.ID, &log.UserID, &log.Operation, &log.InputFile, &log.OutputFile, &log.Details, &log.CreatedAt)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}

func (r *OperationLogRepository) List() ([]models.OperationLog, error) {
	query := `SELECT id, user_id, operation, input_file, output_file, details, created_at FROM operation_logs ORDER BY created_at DESC`
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.OperationLog
	for rows.Next() {
		var log models.OperationLog
		err := rows.Scan(&log.ID, &log.UserID, &log.Operation, &log.InputFile, &log.OutputFile, &log.Details, &log.CreatedAt)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}

func (r *OperationLogRepository) Delete(id uint) error {
	query := `DELETE FROM operation_logs WHERE id = ?`
	_, err := db.DB.Exec(query, id)
	return err
}

func (r *OperationLogRepository) DeleteByUserID(userID uint) error {
	query := `DELETE FROM operation_logs WHERE user_id = ?`
	_, err := db.DB.Exec(query, userID)
	return err
}

var _ models.OperationLogRepository = (*OperationLogRepository)(nil)
