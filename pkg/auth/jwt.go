package auth

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/improwised/datautil/pkg/models"
)

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserExists          = errors.New("user already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidToken        = errors.New("invalid token")
	ErrUnauthorized        = errors.New("unauthorized")
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

var (
	userRepo models.UserRepository
	logRepo models.OperationLogRepository
)

func UserRepository() models.UserRepository {
	return userRepo
}

func SetUserRepository(repo models.UserRepository) {
	userRepo = repo
}

func OperationLogRepository() models.OperationLogRepository {
	return logRepo
}

func SetOperationLogRepository(repo models.OperationLogRepository) {
	logRepo = repo
}

func Register(username, email, password string) (*models.User, error) {
	repo := userRepo
	if repo == nil {
		return nil, fmt.Errorf("user repository not configured")
	}

	existing, err := repo.GetByEmail(email)
	if err == nil && existing != nil {
		return nil, ErrUserExists
	}
	existing, err = repo.GetByUsername(username)
	if err == nil && existing != nil {
		return nil, ErrUserExists
	}

	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Username:  username,
		Email:     email,
		Password:  hashedPassword,
		Role:      "user",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return repo.Create(user)
}

func Login(email, password string) (*models.User, string, error) {
	repo := userRepo
	if repo == nil {
		return nil, "", fmt.Errorf("user repository not configured")
	}

	user, err := repo.GetByEmail(email)
	if err != nil {
		return nil, "", ErrInvalidCredentials
	}

	if !user.Active {
		return nil, "", errors.New("user account is disabled")
	}

	if !CheckPassword(password, user.Password) {
		return nil, "", ErrInvalidCredentials
	}

	jwtObj := NewJWT()
	token, err := jwtObj.GenerateToken(user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, token, nil
}

func GetUserByID(id uint) (*models.User, error) {
	repo := userRepo
	if repo == nil {
		return nil, fmt.Errorf("user repository not configured")
	}

	user, err := repo.GetByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
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
	repo := logRepo
	if repo == nil {
		return
	}

	logEntry := &models.OperationLog{
		UserID:     userID,
		Operation:  operation,
		InputFile:  inputFile,
		OutputFile: outputFile,
		Details:    details,
	}
	_, _ = repo.Create(logEntry)
}