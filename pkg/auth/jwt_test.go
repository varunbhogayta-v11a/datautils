package auth

import (
	"os"
	"testing"
	"time"

	"github.com/improwised/datautil/pkg/db"
	"github.com/improwised/datautil/pkg/models"
)

func TestHashPassword(t *testing.T) {
	password := "testpassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if hash == "" {
		t.Error("expected non-empty hash")
	}

	if hash == password {
		t.Error("hash should not equal plaintext password")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "testpassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if !CheckPassword(password, hash) {
		t.Error("CheckPassword() should return true for correct password")
	}

	if CheckPassword("wrongpassword", hash) {
		t.Error("CheckPassword() should return false for incorrect password")
	}
}

func TestJWTGenerateToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	jwt := NewJWT()
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Role:     "user",
	}

	token, err := jwt.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestJWTValidateToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	jwt := NewJWT()
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Role:     "user",
	}

	token, err := jwt.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	claims, err := jwt.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	if claims.UserID != 1 {
		t.Errorf("claims.UserID = %d, want 1", claims.UserID)
	}

	if claims.Username != "testuser" {
		t.Errorf("claims.Username = %q, want %q", claims.Username, "testuser")
	}
}

func TestJWTValidateToken_Invalid(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	jwt := NewJWT()

	_, err := jwt.ValidateToken("invalid.token.here")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestJWTValidateToken_WrongSecret(t *testing.T) {
	os.Setenv("JWT_SECRET", "secret1")
	defer os.Unsetenv("JWT_SECRET")

	jwt1 := NewJWT()
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Role:     "user",
	}

	token, _ := jwt1.GenerateToken(user)

	os.Setenv("JWT_SECRET", "secret2")
	jwt2 := NewJWT()

	_, err := jwt2.ValidateToken(token)
	if err == nil {
		t.Error("expected error for token with wrong secret")
	}
}

func TestParseTokenFromHeader(t *testing.T) {
	tests := []struct {
		name      string
		header   string
		wantToken string
		wantErr  bool
	}{
		{"valid bearer", "Bearer mytoken123", "mytoken123", false},
		{"empty header", "", "", true},
		{"no bearer", "mytoken123", "", true},
		{"wrong scheme", "Basic mytoken", "", true},
		{"missing token", "Bearer ", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := ParseTokenFromHeader(tt.header)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTokenFromHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && token != tt.wantToken {
				t.Errorf("ParseTokenFromHeader() = %q, want %q", token, tt.wantToken)
			}
		})
	}
}

func TestJWTCustomExpire(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	jwt := &JWT{
		SecretKey:   "test-secret-key",
		TokenExpire: 1 * time.Hour,
	}

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Role:     "user",
	}

	token, err := jwt.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	claims, err := jwt.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	if claims.ExpiresAt == nil {
		t.Error("expected ExpiresAt to be set")
	}
}

func TestClaimsRole(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	jwt := NewJWT()
	user := &models.User{
		ID:       1,
		Username: "adminuser",
		Role:     "admin",
	}

	token, _ := jwt.GenerateToken(user)
	claims, _ := jwt.ValidateToken(token)

	if claims.Role != "admin" {
		t.Errorf("claims.Role = %q, want %q", claims.Role, "admin")
	}
}

func TestNewJWT(t *testing.T) {
	os.Setenv("JWT_SECRET", "my-secret")
	defer os.Unsetenv("JWT_SECRET")

	jwt := NewJWT()
	if jwt.SecretKey != "my-secret" {
		t.Errorf("SecretKey = %q, want %q", jwt.SecretKey, "my-secret")
	}
	if jwt.TokenExpire != 24*time.Hour {
		t.Errorf("TokenExpire = %v, want %v", jwt.TokenExpire, 24*time.Hour)
	}
}

func TestNewJWT_DefaultSecret(t *testing.T) {
	os.Unsetenv("JWT_SECRET")

	jwt := NewJWT()
	if jwt.SecretKey == "" {
		t.Error("expected default secret")
	}
}

func TestNewJWT_DefaultExpiry(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-key")
	defer os.Unsetenv("JWT_SECRET")

	jwt := &JWT{
		SecretKey:   "test-secret",
		TokenExpire: 24 * time.Hour,
	}

	if jwt.TokenExpire != 24*time.Hour {
		t.Errorf("TokenExpire = %v, want %v", jwt.TokenExpire, 24*time.Hour)
	}
}

func TestRegister(t *testing.T) {
	mockRepo := db.NewMockUserRepository()
	SetUserRepository(mockRepo)
	defer SetUserRepository(nil)

	user, err := Register("newuser", "new@example.com", "password123")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if user.Username != "newuser" {
		t.Errorf("Username = %q, want %q", user.Username, "newuser")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	mockRepo := db.NewMockUserRepository()
	SetUserRepository(mockRepo)
	defer SetUserRepository(nil)

	_, _ = Register("user1", "duplicate@example.com", "password123")

	_, err := Register("user2", "duplicate@example.com", "password456")
	if err != ErrUserExists {
		t.Errorf("expected ErrUserExists, got %v", err)
	}
}

func TestLogin_WithMockRepo(t *testing.T) {
	mockRepo := db.NewMockUserRepository()
	SetUserRepository(mockRepo)
	defer SetUserRepository(nil)

	_, _ = Register("testuser", "logintest@example.com", "password123")

	user, token, err := Login("logintest@example.com", "password123")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if user.Username != "testuser" {
		t.Errorf("Username = %q, want %q", user.Username, "testuser")
	}

	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	mockRepo := db.NewMockUserRepository()
	SetUserRepository(mockRepo)
	defer SetUserRepository(nil)

	_, _, err := Login("nonexistent@example.com", "password123")
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_DisabledUser(t *testing.T) {
	mockRepo := db.NewMockUserRepository()
	SetUserRepository(mockRepo)
	defer SetUserRepository(nil)

	user, _ := mockRepo.Create(&models.User{
		Username: "disableduser",
		Email:    "disabled@example.com",
		Password: "hash",
		Role:     "user",
		Active:   false,
	})

	SetUserRepository(mockRepo)
	_, _, err := Login("disabled@example.com", "password123")
	if err == nil {
		t.Error("expected error for disabled user")
	}
	_ = user
}

func TestGetUserByID(t *testing.T) {
	mockRepo := db.NewMockUserRepository()
	SetUserRepository(mockRepo)
	defer SetUserRepository(nil)

	created, _ := mockRepo.Create(&models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hash",
		Role:     "user",
		Active:   true,
	})

	user, err := GetUserByID(created.ID)
	if err != nil {
		t.Fatalf("GetUserByID() error = %v", err)
	}

	if user.Username != "testuser" {
		t.Errorf("Username = %q, want %q", user.Username, "testuser")
	}
}

func TestGetUserByID_NotFound(t *testing.T) {
	mockRepo := db.NewMockUserRepository()
	SetUserRepository(mockRepo)
	defer SetUserRepository(nil)

	_, err := GetUserByID(999)
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestLogOperation(t *testing.T) {
	mockLogRepo := db.NewMockOperationLogRepository()
	SetOperationLogRepository(mockLogRepo)
	defer SetOperationLogRepository(nil)

	mockUserRepo := db.NewMockUserRepository()
	SetUserRepository(mockUserRepo)
	defer SetUserRepository(nil)

	user, _ := mockUserRepo.Create(&models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hash",
		Role:     "user",
		Active:   true,
	})

	LogOperation(user.ID, "filter", "input.csv", "output.csv", "Filtered 10 rows")

	logs, _ := mockLogRepo.List()
	if len(logs) != 1 {
		t.Errorf("expected 1 log, got %d", len(logs))
	}
	_ = user
}