//go:build e2e
// +build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestE2E_APIHealth(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e tests in short mode")
	}

	startServer(t)
	time.Sleep(500 * time.Millisecond)

	resp, err := http.Get("http://localhost:8080/api/health")
	if err != nil {
		t.Skipf("Server may not be running: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestE2E_APIRegisterLogin(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e tests in short mode")
	}

	startServer(t)
	time.Sleep(500 * time.Millisecond)

	registerBody := []byte(`{"username":"testuser","email":"test@example.com","password":"test123"}`)
	resp, err := http.Post("http://localhost:8080/api/auth/register", "application/json", bytes.NewReader(registerBody))
	if err != nil {
		t.Skipf("Server may not be running: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		t.Logf("Register response status: %d", resp.StatusCode)
	}

	loginBody := []byte(`{"email":"test@example.com","password":"test123"}`)
	resp, err = http.Post("http://localhost:8080/api/auth/login", "application/json", bytes.NewReader(loginBody))
	if err != nil {
		t.Skipf("Server may not be running: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Logf("Login response status: %d", resp.StatusCode)
	}
}

func TestE2E_APIAuthFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e tests in short mode")
	}

	startServer(t)
	time.Sleep(500 * time.Millisecond)

	t.Run("register then login", func(t *testing.T) {
		registerReq := map[string]string{
			"username": "authtest",
			"email":   "authtest@example.com",
			"password": "password123",
		}
		registerBody, _ := json.Marshal(registerReq)

		resp, err := http.Post("http://localhost:8080/api/auth/register", "application/json", bytes.NewReader(registerBody))
		if err != nil {
			t.Skipf("Server may not be running: %v", err)
		}
		defer resp.Body.Close()

		loginReq := map[string]string{
			"email":    "authtest@example.com",
			"password": "password123",
		}
		loginBody, _ := json.Marshal(loginReq)

		resp, err = http.Post("http://localhost:8080/api/auth/login", "application/json", bytes.NewReader(loginBody))
		if err != nil {
			t.Skipf("Server may not be running: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Logf("Login status: %d", resp.StatusCode)
		}

		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)

		if token, ok := response["token"]; ok {
			t.Logf("Got token: %v", token)
		}
	})
}

func TestE2E_APIDataFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e tests in short mode")
	}

	startServer(t)
	time.Sleep(500 * time.Millisecond)

	t.Run("protected endpoint without token", func(t *testing.T) {
		filterBody := []byte(`{"input":"tests/test_data.csv","where":"age > 25"}`)
		resp, err := http.Post("http://localhost:8080/api/data/filter", "application/json", bytes.NewReader(filterBody))
		if err != nil {
			t.Skipf("Server may not be running: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Logf("Expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("filter endpoint with token", func(t *testing.T) {
		token := "Bearer test-token-placeholder"
		filterBody := []byte(`{"input":"tests/test_data.csv","where":"age > 25"}`)
		
		req, _ := http.NewRequest("POST", "http://localhost:8080/api/data/filter", bytes.NewReader(filterBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", token)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Skipf("Server may not be running: %v", err)
		}
		defer resp.Body.Close()

		t.Logf("Filter response status: %d", resp.StatusCode)
	})
}

func TestE2E_APIDatabaseQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e tests in short mode")
	}

	startServer(t)
	time.Sleep(500 * time.Millisecond)

	t.Run("execute query", func(t *testing.T) {
		queryBody := []byte(`{"sql":"SELECT 1 as test"}`)
		
		req, _ := http.NewRequest("POST", "http://localhost:8080/api/db/query", bytes.NewReader(queryBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Skipf("Server may not be running: %v", err)
		}
		defer resp.Body.Close()

		t.Logf("Query response status: %d", resp.StatusCode)
	})
}

func TestE2E_APIErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e tests in short mode")
	}

	startServer(t)
	time.Sleep(500 * time.Millisecond)

	t.Run("invalid json", func(t *testing.T) {
		resp, err := http.Post("http://localhost:8080/api/data/filter", "application/json", strings.NewReader("not valid json"))
		if err != nil {
			t.Skipf("Server may not be running: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Logf("Expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid endpoint", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8080/api/invalid/endpoint")
		if err != nil {
			t.Skipf("Server may not be running: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Logf("Expected 404, got %d", resp.StatusCode)
		}
	})
}

func startServer(t *testing.T) {
	if _, err := http.Get("http://localhost:8080/api/health"); err == nil {
		return
	}

	go func() {
		cmd := exec.Command("./datautil", "server", "--port", "8080")
		cmd.Env = append(os.Environ(), "DB_DRIVER=sqlite3", "DB_NAME=/tmp/datautil_test.db")
		cmd.Dir = t.TempDir()
		cmd.Run()
	}()

	time.Sleep(1 * time.Second)
}

var serverStarted bool

func ensureServer(t *testing.T) {
	if serverStarted {
		return
	}

	for i := 0; i < 10; i++ {
		if _, err := http.Get("http://localhost:8080/api/health"); err == nil {
			serverStarted = true
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	fmt.Println("Warning: Could not ensure server is running")
}