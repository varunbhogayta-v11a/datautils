package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/varunbhogayta-v11a/datautils/pkg/auth"
	"github.com/varunbhogayta-v11a/datautils/pkg/db"
	"github.com/varunbhogayta-v11a/datautils/pkg/models"
)

type QueryRequest struct {
	SQL   string `json:"sql"`
	Limit int    `json:"limit,omitempty"`
}

func HandleAPIQuery(w http.ResponseWriter, r *http.Request) {
	claims, err := authenticateRequest(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, APIResponse{Success: false, Error: err.Error()})
		return
	}

	tempUser := &models.User{Role: claims.Role}
	if !tempUser.HasPermission("read") {
		writeJSON(w, http.StatusForbidden, APIResponse{Success: false, Error: "Permission denied"})
		return
	}

	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Success: false, Error: "Method not allowed"})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "Invalid request body"})
		return
	}
	defer r.Body.Close()

	var req QueryRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "Invalid JSON"})
		return
	}

	sqlQuery := strings.TrimSpace(req.SQL)
	if sqlQuery == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "sql is required"})
		return
	}

	upperQuery := strings.ToUpper(sqlQuery)
	if !strings.HasPrefix(upperQuery, "SELECT") && !strings.HasPrefix(upperQuery, "PRAGMA") && !strings.HasPrefix(upperQuery, "SHOW") {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "Only SELECT queries allowed"})
		return
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}
	sqlQuery += fmt.Sprintf(" LIMIT %d", limit)

	rows, err := db.DB.Query(sqlQuery)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: fmt.Sprintf("Query error: %v", err)})
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: fmt.Sprintf("Column error: %v", err)})
		return
	}

	resultRows := make([][]string, 0)
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	rowCount := 0
	for rows.Next() {
		err := rows.Scan(valuePtrs...)
		if err != nil {
			continue
		}

		rowValues := make([]string, len(columns))
		for i, v := range values {
			if v == nil {
				rowValues[i] = "NULL"
			} else {
				rowValues[i] = fmt.Sprintf("%v", v)
			}
		}
		resultRows = append(resultRows, rowValues)
		rowCount++
	}

	auth.LogOperation(claims.UserID, "query", sqlQuery, "", "")

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"columns": columns,
			"rows":    resultRows,
			"count":   rowCount,
		},
		Message: fmt.Sprintf("%d rows", rowCount),
	})
}

type InsertRequest struct {
	Table  string `json:"table"`
	Values string `json:"values"`
}

func HandleAPIInsert(w http.ResponseWriter, r *http.Request) {
	claims, err := authenticateRequest(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, APIResponse{Success: false, Error: err.Error()})
		return
	}

	tempUser := &models.User{Role: claims.Role}
	if !tempUser.HasPermission("write") {
		writeJSON(w, http.StatusForbidden, APIResponse{Success: false, Error: "Permission denied - write access required"})
		return
	}

	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Success: false, Error: "Method not allowed"})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "Invalid request body"})
		return
	}
	defer r.Body.Close()

	var req InsertRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "Invalid JSON"})
		return
	}

	if req.Table == "" || req.Values == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "table and values are required"})
		return
	}

	pairs := strings.Split(req.Values, ",")
	cols := make([]string, 0, len(pairs))
	vals := make([]string, 0, len(pairs))
	argsSlice := make([]interface{}, 0, len(pairs))

	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}
		cols = append(cols, strings.TrimSpace(parts[0]))
		vals = append(vals, "?")
		argsSlice = append(argsSlice, strings.TrimSpace(parts[1]))
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		req.Table, strings.Join(cols, ", "), strings.Join(vals, ", "))

	result, err := db.DB.Exec(query, argsSlice...)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: fmt.Sprintf("Insert error: %v", err)})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	auth.LogOperation(claims.UserID, "insert", req.Table, "", req.Values)

	writeJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"table":        req.Table,
			"rowsAffected": rowsAffected,
		},
		Message: fmt.Sprintf("Inserted %d row(s)", rowsAffected),
	})
}

type UpdateRequest struct {
	Table string `json:"table"`
	Set   string `json:"set"`
	Where string `json:"where,omitempty"`
}

func HandleAPIUpdate(w http.ResponseWriter, r *http.Request) {
	claims, err := authenticateRequest(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, APIResponse{Success: false, Error: err.Error()})
		return
	}

	tempUser := &models.User{Role: claims.Role}
	if !tempUser.HasPermission("write") {
		writeJSON(w, http.StatusForbidden, APIResponse{Success: false, Error: "Permission denied - write access required"})
		return
	}

	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Success: false, Error: "Method not allowed"})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "Invalid request body"})
		return
	}
	defer r.Body.Close()

	var req UpdateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "Invalid JSON"})
		return
	}

	if req.Table == "" || req.Set == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "table and set are required"})
		return
	}

	query := fmt.Sprintf("UPDATE %s SET %s", req.Table, req.Set)
	if req.Where != "" {
		query += " WHERE " + req.Where
	}

	result, err := db.DB.Exec(query)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: fmt.Sprintf("Update error: %v", err)})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	auth.LogOperation(claims.UserID, "update", req.Table, "", req.Set+" "+req.Where)

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"table":        req.Table,
			"rowsAffected": rowsAffected,
		},
		Message: fmt.Sprintf("Updated %d row(s)", rowsAffected),
	})
}

type DeleteRequest struct {
	Table string `json:"table"`
	Where string `json:"where,omitempty"`
}

func HandleAPIDelete(w http.ResponseWriter, r *http.Request) {
	claims, err := authenticateRequest(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, APIResponse{Success: false, Error: err.Error()})
		return
	}

	tempUser := &models.User{Role: claims.Role}
	if !tempUser.HasPermission("delete") {
		writeJSON(w, http.StatusForbidden, APIResponse{Success: false, Error: "Permission denied - delete access required"})
		return
	}

	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Success: false, Error: "Method not allowed"})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "Invalid request body"})
		return
	}
	defer r.Body.Close()

	var req DeleteRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "Invalid JSON"})
		return
	}

	if req.Table == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "table is required"})
		return
	}

	query := fmt.Sprintf("DELETE FROM %s", req.Table)
	if req.Where != "" {
		query += " WHERE " + req.Where
	}

	result, err := db.DB.Exec(query)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: fmt.Sprintf("Delete error: %v", err)})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	auth.LogOperation(claims.UserID, "delete", req.Table, "", req.Where)

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"table":        req.Table,
			"rowsAffected": rowsAffected,
		},
		Message: fmt.Sprintf("Deleted %d row(s)", rowsAffected),
	})
}

func HandleAPIUsers(w http.ResponseWriter, r *http.Request) {
	_, err := authenticateRequest(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, APIResponse{Success: false, Error: err.Error()})
		return
	}

	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Success: false, Error: "Method not allowed"})
		return
	}

	rows, err := db.DB.Query("SELECT id, username, email, role, active FROM users")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: err.Error()})
		return
	}
	defer rows.Close()

	users := make([]map[string]interface{}, 0)
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.Active); err != nil {
			continue
		}
		users = append(users, map[string]interface{}{
			"id":       u.ID,
			"username": u.Username,
			"email":    u.Email,
			"role":     u.Role,
			"active":   u.Active,
		})
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"users": users,
			"count": len(users),
		},
	})
}

func HandleAPILogs(w http.ResponseWriter, r *http.Request) {
	claims, err := authenticateRequest(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, APIResponse{Success: false, Error: err.Error()})
		return
	}

	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Success: false, Error: "Method not allowed"})
		return
	}

	query := "SELECT id, user_id, operation, input_file, output_file, details, created_at FROM operation_logs ORDER BY created_at DESC LIMIT 50"
	if claims.Role != "admin" {
		query = fmt.Sprintf("SELECT id, user_id, operation, input_file, output_file, details, created_at FROM operation_logs WHERE user_id = %d ORDER BY created_at DESC LIMIT 50", claims.UserID)
	}

	rows, err := db.DB.Query(query)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: err.Error()})
		return
	}
	defer rows.Close()

	logs := make([]map[string]interface{}, 0)
	for rows.Next() {
		var l models.OperationLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.Operation, &l.InputFile, &l.OutputFile, &l.Details, &l.CreatedAt); err != nil {
			continue
		}
		logs = append(logs, map[string]interface{}{
			"id":          l.ID,
			"user_id":     l.UserID,
			"operation":   l.Operation,
			"input_file":  l.InputFile,
			"output_file": l.OutputFile,
			"details":     l.Details,
			"created_at":  l.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"logs":  logs,
			"count": len(logs),
		},
	})
}
