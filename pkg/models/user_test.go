package models

import (
	"testing"
)

func TestUserHasPermission(t *testing.T) {
	tests := []struct {
		name   string
		role   string
		action string
		want   bool
	}{
		{"admin can delete", "admin", "delete", true},
		{"admin can read", "admin", "read", true},
		{"admin can admin", "admin", "admin", true},
		{"user cannot delete", "user", "delete", false},
		{"user can read", "user", "read", true},
		{"user cannot admin", "user", "admin", false},
		{"guest can read", "guest", "read", true},
		{"guest cannot write", "guest", "write", false},
		{"guest cannot delete", "guest", "delete", false},
		{"unknown role", "unknown", "read", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Role: tt.role}
			if got := user.HasPermission(tt.action); got != tt.want {
				t.Errorf("User.HasPermission(%q) = %v, want %v", tt.action, got, tt.want)
			}
		})
	}
}

func TestRoleConstants(t *testing.T) {
	if RoleAdmin != "admin" {
		t.Errorf("RoleAdmin = %q, want %q", RoleAdmin, "admin")
	}
	if RoleUser != "user" {
		t.Errorf("RoleUser = %q, want %q", RoleUser, "user")
	}
	if RoleGuest != "guest" {
		t.Errorf("RoleGuest = %q, want %q", RoleGuest, "guest")
	}
}

func TestGetTableName(t *testing.T) {
	t.Run("postgres", func(t *testing.T) {
		tables := GetTableName("postgres")
		if tables["users"] == "" {
			t.Error("expected users table definition for postgres")
		}
		if tables["operation_logs"] == "" {
			t.Error("expected operation_logs table definition for postgres")
		}
	})

	t.Run("sqlite", func(t *testing.T) {
		tables := GetTableName("sqlite")
		if tables["users"] == "" {
			t.Error("expected users table definition for sqlite")
		}
	})

	t.Run("mysql", func(t *testing.T) {
		tables := GetTableName("mysql")
		if tables["users"] == "" {
			t.Error("expected users table definition for mysql")
		}
	})

	t.Run("unsupported", func(t *testing.T) {
		tables := GetTableName("unsupported")
		if len(tables) != 0 {
			t.Error("expected empty tables for unsupported driver")
		}
	})
}

func TestRouteConfigGetDynamicParams(t *testing.T) {
	t.Run("nil params", func(t *testing.T) {
		route := &RouteConfig{}
		params := route.GetDynamicParams()
		if params != nil {
			t.Error("expected nil for nil DynamicParams")
		}
	})

	t.Run("valid params", func(t *testing.T) {
		route := &RouteConfig{
			DynamicParams: []byte(`[{"name":"id","type":"int","default":"1","required":true}]`),
		}
		params := route.GetDynamicParams()
		if len(params) != 1 {
			t.Fatalf("expected 1 param, got %d", len(params))
		}
		if params[0].Name != "id" {
			t.Errorf("param name = %q, want %q", params[0].Name, "id")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		route := &RouteConfig{
			DynamicParams: []byte(`invalid`),
		}
		params := route.GetDynamicParams()
		if params != nil {
			t.Error("expected nil for invalid JSON")
		}
	})
}

func TestRouteConfigSetDynamicParams(t *testing.T) {
	t.Run("valid params", func(t *testing.T) {
		route := &RouteConfig{}
		params := []DynamicParam{
			{Name: "id", Type: "int", Default: "1", Required: true},
		}
		err := route.SetDynamicParams(params)
		if err != nil {
			t.Fatalf("SetDynamicParams() error = %v", err)
		}
		if route.DynamicParams == nil {
			t.Error("expected DynamicParams to be set")
		}
	})
}

func TestPaginationFields(t *testing.T) {
	p := &Pagination{
		Page:       1,
		Limit:      10,
		TotalRows:  100,
		TotalPages: 10,
		HasNext:    true,
		HasPrev:    false,
	}

	if p.Page != 1 {
		t.Errorf("Page = %d, want 1", p.Page)
	}
	if p.Limit != 10 {
		t.Errorf("Limit = %d, want 10", p.Limit)
	}
	if p.TotalRows != 100 {
		t.Errorf("TotalRows = %d, want 100", p.TotalRows)
	}
	if p.TotalPages != 10 {
		t.Errorf("TotalPages = %d, want 10", p.TotalPages)
	}
}

func TestRouteResponseFields(t *testing.T) {
	r := &RouteResponse{
		Success: true,
		Data:    "test data",
		Error:   "",
		Message: "success",
	}

	if !r.Success {
		t.Error("expected Success = true")
	}
	if r.Message != "success" {
		t.Errorf("Message = %q, want %q", r.Message, "success")
	}
}

func TestRouteResponseWithPagination(t *testing.T) {
	p := &Pagination{Page: 1, Limit: 10, TotalRows: 50}
	r := &RouteResponse{
		Success:    true,
		Data:       []string{"a", "b"},
		Pagination: p,
	}

	if r.Pagination == nil {
		t.Error("expected Pagination to be set")
	}
	if r.Pagination.TotalRows != 50 {
		t.Errorf("TotalRows = %d, want 50", r.Pagination.TotalRows)
	}
}
