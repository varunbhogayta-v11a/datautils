package auth

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/improwised/datautil/pkg/db"
	"github.com/improwised/datautil/pkg/models"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidToken       = errors.New("invalid token")
	ErrUnauthorized       = errors.New("unauthorized")
)

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type JWT struct {
	SecretKey     string
	TokenExpire   time.Duration
	RefreshExpire time.Duration
}

func NewJWT() *JWT {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "datautil-secret-key-change-in-production"
	}
	return &JWT{
		SecretKey:     secret,
		TokenExpire:   24 * time.Hour,
		RefreshExpire: 7 * 24 * time.Hour,
	}
}

func (j *JWT) GenerateToken(user *models.User) (string, error) {
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.TokenExpire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "datautil",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.SecretKey))
}

func (j *JWT) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.SecretKey), nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func Register(username, email, password string) (*models.User, error) {
	var count int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ? OR username = ?", email, username).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrUserExists
	}

	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	result, err := db.DB.Exec(
		"INSERT INTO users (username, email, password, role, active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))",
		username, email, hashedPassword, "user", true,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	id, _ := result.LastInsertId()
	return &models.User{
		ID:        uint(id),
		Username:  username,
		Email:     email,
		Password:  hashedPassword,
		Role:      "user",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func Login(email, password string) (*models.User, string, error) {
	var user models.User
	err := db.DB.QueryRow(
		"SELECT id, username, email, password, role, active, created_at, updated_at FROM users WHERE email = ?",
		email,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Role, &user.Active, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}

	if !user.Active {
		return nil, "", errors.New("user account is disabled")
	}

	if !CheckPassword(password, user.Password) {
		return nil, "", ErrInvalidCredentials
	}

	jwt := NewJWT()
	token, err := jwt.GenerateToken(&user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return &user, token, nil
}

func GetUserByID(id uint) (*models.User, error) {
	var user models.User
	err := db.DB.QueryRow(
		"SELECT id, username, email, password, role, active, created_at, updated_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Role, &user.Active, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func ParseTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", ErrUnauthorized
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", ErrInvalidToken
	}

	return parts[1], nil
}

func LogOperation(userID uint, operation, inputFile, outputFile, details string) {
	_, _ = db.DB.Exec(
		"INSERT INTO operation_logs (user_id, operation, input_file, output_file, details, created_at) VALUES (?, ?, ?, ?, ?, datetime('now'))",
		userID, operation, inputFile, outputFile, details,
	)
}
