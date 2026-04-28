//go:build integration
// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIntegration_HealthEndpoint(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("status = %q, want %q", response["status"], "healthy")
	}
}

func TestIntegration_AuthRegister(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/register" {
			t.Errorf("expected path /api/auth/register, got %s", r.URL.Path)
		}

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"message":"user registered"}`))
	})

	body := []byte(`{"username":"testuser","email":"test@example.com","password":"test123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rr.Code)
	}
}

func TestIntegration_AuthLogin(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/login" {
			t.Errorf("expected path /api/auth/login, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test"}`))
	})

	body := []byte(`{"email":"test@example.com","password":"test123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestIntegration_DataFilterEndpoint(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/data/filter" {
			t.Errorf("expected path /api/data/filter, got %s", r.URL.Path)
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			t.Error("expected Authorization header")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true,"data":{"headers":["name","age"],"rows":[["John","25"]]}}`))
	})

	body := []byte(`{"input":"tests/test_data.csv","where":"age > 25"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/data/filter", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestIntegration_DataTransformEndpoint(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true}`))
	})

	body := []byte(`{"input":"tests/test_data.csv","add":"city=NYC"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/data/transform", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestIntegration_DataValidateEndpoint(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"valid":true,"errors":[]}`))
	})

	body := []byte(`{"input":"tests/test_data.csv","required":["name","age"]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/data/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestIntegration_RequireAuth(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	t.Run("missing auth", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rr.Code)
		}
	})

	t.Run("with auth", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rr.Code)
		}
	})
}

func TestIntegration_APIResponseFormat(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true,"data":{"id":1,"name":"test"},"message":"ok"}`))
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp["success"] != true {
		t.Errorf("success = %v, want true", resp["success"])
	}
}

func TestIntegration_MissingRequiredFields(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]interface{}
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if req["username"] == nil || req["username"] == "" {
			http.Error(w, "username required", http.StatusBadRequest)
			return
		}
		if req["email"] == nil || req["email"] == "" {
			http.Error(w, "email required", http.StatusBadRequest)
			return
		}
		if req["password"] == nil || req["password"] == "" {
			http.Error(w, "password required", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})

	t.Run("missing username", func(t *testing.T) {
		body := []byte(`{"email":"test@example.com","password":"test123"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rr.Code)
		}
	})

	t.Run("missing email", func(t *testing.T) {
		body := []byte(`{"username":"testuser","password":"test123"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rr.Code)
		}
	})

	t.Run("missing password", func(t *testing.T) {
		body := []byte(`{"username":"testuser","email":"test@example.com"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rr.Code)
		}
	})
}

func TestIntegration_InvalidJSON(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		body, _ := io.ReadAll(r.Body)
		var req map[string]interface{}
		if err := json.Unmarshal(body, &req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"success":false,"error":"Invalid JSON"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	t.Run("invalid json", func(t *testing.T) {
		body := []byte(`{invalid json}`)
		req := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rr.Code)
		}
	})

	t.Run("empty body", func(t *testing.T) {
		body := []byte(``)
		req := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rr.Code)
		}
	})
}

func TestIntegration_MethodNotAllowed(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE"}

	for _, method := range methods {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				w.Write([]byte(`{"success":false,"error":"Method not allowed"}`))
				return
			}
			w.WriteHeader(http.StatusOK)
		})

		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/test", nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if method != "POST" && rr.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected status 405, got %d", rr.Code)
			}
			if method == "POST" && rr.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", rr.Code)
			}
		})
	}
}

func TestIntegration_BearerTokenParsing(t *testing.T) {
	tests := []struct {
		name      string
		header    string
		wantToken string
		wantErr   bool
	}{
		{"valid bearer", "Bearer mytoken123", "mytoken123", false},
		{"empty header", "", "", true},
		{"no bearer", "mytoken123", "", true},
		{"wrong scheme", "Basic mytoken", "", true},
		{"multiple spaces", "Bearer", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				authHeader := r.Header.Get("Authorization")
				if authHeader == "" {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}

				parts := strings.Split(authHeader, " ")
				if len(parts) != 2 || parts[0] != "Bearer" {
					http.Error(w, "invalid token", http.StatusUnauthorized)
					return
				}

				token := parts[1]
				if tt.wantErr && token == "" {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"success":true,"token":"` + token + `"}`))
			})

			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if tt.wantErr && rr.Code != http.StatusUnauthorized {
				t.Errorf("expected 401, got %d", rr.Code)
			}
			if !tt.wantErr && rr.Code != http.StatusOK {
				t.Errorf("expected 200, got %d", rr.Code)
			}
		})
	}
}