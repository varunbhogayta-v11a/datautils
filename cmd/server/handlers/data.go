package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/varunbhogayta-v11a/datautils/pkg/auth"
	"github.com/varunbhogayta-v11a/datautils/pkg/data"
	"github.com/varunbhogayta-v11a/datautils/pkg/db"
	"github.com/varunbhogayta-v11a/datautils/pkg/models"
	"github.com/varunbhogayta-v11a/datautils/pkg/operations"
)

type FilterRequest struct {
	Input  string `json:"input"`
	Where  string `json:"where,omitempty"`
	Select string `json:"select,omitempty"`
	Invert bool   `json:"invert,omitempty"`
	Output string `json:"output,omitempty"`
	Format string `json:"format,omitempty"`
}

func HandleAPIFilter(w http.ResponseWriter, r *http.Request) {
	claims, err := authenticateRequest(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, APIResponse{Success: false, Error: err.Error()})
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

	var req FilterRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "Invalid JSON"})
		return
	}

	if req.Input == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "input file is required"})
		return
	}

	ds, err := data.ReadFile(req.Input)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: fmt.Sprintf("Failed to read file: %v", err)})
		return
	}

	if req.Where != "" {
		ds, err = operations.FilterRows(ds, req.Where, req.Invert)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: fmt.Sprintf("Filter error: %v", err)})
			return
		}
	}

	if req.Select != "" {
		ds = operations.SelectColumns(ds, strings.Split(req.Select, ","))
	}

	if req.Output != "" {
		format := req.Format
		if format == "" {
			format = data.DetectFormat(req.Output)
		}
		if err := data.WriteFile(ds, req.Output, format); err != nil {
			writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: fmt.Sprintf("Write error: %v", err)})
			return
		}
		auth.LogOperation(claims.UserID, "filter", req.Input, req.Output, req.Where)
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"headers": ds.Headers,
			"rows":    ds.Rows,
			"count":   ds.RowCount(),
		},
		Message: fmt.Sprintf("Filtered %d rows", ds.RowCount()),
	})
}

type TransformRequest struct {
	Input  string `json:"input"`
	Add    string `json:"add,omitempty"`
	Remove string `json:"remove,omitempty"`
	Rename string `json:"rename,omitempty"`
	Output string `json:"output,omitempty"`
	Format string `json:"format,omitempty"`
}

func HandleAPITransform(w http.ResponseWriter, r *http.Request) {
	claims, err := authenticateRequest(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, APIResponse{Success: false, Error: err.Error()})
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

	var req TransformRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "Invalid JSON"})
		return
	}

	if req.Input == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "input file is required"})
		return
	}

	ds, err := data.ReadFile(req.Input)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: fmt.Sprintf("Failed to read file: %v", err)})
		return
	}

	if req.Add != "" {
		parts := strings.SplitN(req.Add, "=", 2)
		if len(parts) == 2 {
			ds = operations.AddColumn(ds, strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}

	if req.Remove != "" {
		ds = operations.RemoveColumns(ds, strings.Split(req.Remove, ","))
	}

	if req.Rename != "" {
		parts := strings.SplitN(req.Rename, ":", 2)
		if len(parts) == 2 {
			ds = operations.RenameColumn(ds, strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}

	if req.Output != "" {
		format := req.Format
		if format == "" {
			format = data.DetectFormat(req.Output)
		}
		if err := data.WriteFile(ds, req.Output, format); err != nil {
			writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: fmt.Sprintf("Write error: %v", err)})
			return
		}
		auth.LogOperation(claims.UserID, "transform", req.Input, req.Output, req.Add+req.Remove+req.Rename)
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"headers": ds.Headers,
			"rows":    ds.Rows,
			"count":   ds.RowCount(),
		},
		Message: "Transform completed",
	})
}

type ValidateRequest struct {
	Input    string `json:"input"`
	Required string `json:"required,omitempty"`
	Types    string `json:"types,omitempty"`
}

func HandleAPIValidate(w http.ResponseWriter, r *http.Request) {
	claims, err := authenticateRequest(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, APIResponse{Success: false, Error: err.Error()})
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

	var req ValidateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "Invalid JSON"})
		return
	}

	if req.Input == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "input file is required"})
		return
	}

	ds, err := data.ReadFile(req.Input)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: fmt.Sprintf("Failed to read file: %v", err)})
		return
	}

	var required []string
	if req.Required != "" {
		required = strings.Split(req.Required, ",")
	}

	typesMap := make(map[string]string)
	if req.Types != "" {
		for _, t := range strings.Split(req.Types, ",") {
			parts := strings.SplitN(t, ":", 2)
			if len(parts) == 2 {
				typesMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	result := operations.ValidateDataset(ds, required, typesMap)

	if result.Valid {
		auth.LogOperation(claims.UserID, "validate", req.Input, "", req.Required+req.Types)
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: result.Valid,
		Data: map[string]interface{}{
			"valid":   result.Valid,
			"errors":  result.Errors,
			"rows":    ds.RowCount(),
			"columns": ds.ColCount(),
		},
		Message: "Validation completed",
	})
}

type ExportRequest struct {
	Input  string `json:"input"`
	Output string `json:"output,omitempty"`
	To     string `json:"to"`
	Pretty bool   `json:"pretty,omitempty"`
	Stats  bool   `json:"stats,omitempty"`
}

func HandleAPIExport(w http.ResponseWriter, r *http.Request) {
	claims, err := authenticateRequest(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, APIResponse{Success: false, Error: err.Error()})
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

	var req ExportRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "Invalid JSON"})
		return
	}

	if req.Input == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "input file is required"})
		return
	}

	ds, err := data.ReadFile(req.Input)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: fmt.Sprintf("Failed to read file: %v", err)})
		return
	}

	respData := map[string]interface{}{
		"headers": ds.Headers,
		"rows":    ds.Rows,
		"count":   ds.RowCount(),
	}

	if req.Stats {
		respData["statistics"] = map[string]interface{}{
			"rows":    ds.RowCount(),
			"columns": ds.ColCount(),
			"headers": ds.Headers,
		}
	}

	if req.Output != "" {
		if err := data.WriteFile(ds, req.Output, req.To); err != nil {
			writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: fmt.Sprintf("Write error: %v", err)})
			return
		}
		auth.LogOperation(claims.UserID, "export", req.Input, req.Output, "")
		respData["output"] = req.Output
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    respData,
		Message: "Export completed",
	})
}

type ImportRequest struct {
	Source      string `json:"source"`
	Table       string `json:"table"`
	CreateTable bool   `json:"create,omitempty"`
	IfNotExists bool   `json:"ifNotExists,omitempty"`
	Truncate    bool   `json:"truncate,omitempty"`
}

func HandleAPIImport(w http.ResponseWriter, r *http.Request) {
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

	var req ImportRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "Invalid JSON"})
		return
	}

	if req.Source == "" || req.Table == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "source and table are required"})
		return
	}

	ds, err := data.ReadFile(req.Source)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: fmt.Sprintf("Failed to read file: %v", err)})
		return
	}

	if req.CreateTable {
		if err := createTableFromDataset(req.Table, ds.Headers); err != nil {
			writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: fmt.Sprintf("Create table error: %v", err)})
			return
		}
	}

	batchSize := 1000
	totalInserted := 0

	for i := 0; i < len(ds.Rows); i += batchSize {
		end := i + batchSize
		if end > len(ds.Rows) {
			end = len(ds.Rows)
		}

		columns := strings.Join(ds.Headers, ", ")
		placeholders := make([]string, len(ds.Headers))
		for j := range ds.Headers {
			placeholders[j] = "?"
		}

		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES ", req.Table, columns)
		values := make([]interface{}, 0, len(ds.Headers)*(end-i))

		for rowIdx := i; rowIdx < end; rowIdx++ {
			rowValues := make([]string, len(ds.Headers))
			for j := range ds.Headers {
				rowValues[j] = "?"
				if j < len(ds.Rows[rowIdx]) {
					values = append(values, ds.Rows[rowIdx][j])
				} else {
					values = append(values, nil)
				}
			}
			query += fmt.Sprintf("(%s), ", strings.Join(rowValues, ", "))
		}
		query = strings.TrimSuffix(query, ", ") + ";"

		_, err = db.DB.Exec(query, values...)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: fmt.Sprintf("Insert error: %v", err)})
			return
		}

		totalInserted += end - i
	}

	auth.LogOperation(claims.UserID, "import", req.Source, req.Table, fmt.Sprintf("%d rows", totalInserted))

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"table":  req.Table,
			"rows":   totalInserted,
			"source": req.Source,
		},
		Message: fmt.Sprintf("Successfully imported %d rows", totalInserted),
	})
}

func createTableFromDataset(tableName string, headers []string) error {
	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INTEGER PRIMARY KEY AUTOINCREMENT", tableName)
	for _, col := range headers {
		safeCol := sanitizeColumnName(col)
		sql += fmt.Sprintf(", %s TEXT", safeCol)
	}
	sql += ")"
	_, err := db.DB.Exec(sql)
	return err
}

func sanitizeColumnName(name string) string {
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	result := ""
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
			result += string(c)
		}
	}
	if result == "" {
		return "column"
	}
	return result
}

const uploadDir = "./data/uploads"

func HandleUpload(w http.ResponseWriter, r *http.Request) {
	claims, err := authenticateRequest(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, APIResponse{Success: false, Error: err.Error()})
		return
	}

	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Success: false, Error: "Method not allowed"})
		return
	}

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		writeJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Error: "Failed to create upload directory"})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "No file uploaded"})
		return
	}
	defer file.Close()

	destPath := uploadDir + "/" + header.Filename
	dest, err := os.Create(destPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Error: "Failed to save file"})
		return
	}
	defer dest.Close()

	if _, err := io.Copy(dest, file); err != nil {
		writeJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Error: "Failed to save file"})
		return
	}

	auth.LogOperation(claims.UserID, "upload", header.Filename, "", fmt.Sprintf("%.2f KB", float64(header.Size)/1024))

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"filename": header.Filename,
			"size":     header.Size,
			"path":     destPath,
		},
		Message: "File uploaded successfully",
	})
}

func HandleListFiles(w http.ResponseWriter, r *http.Request) {
	_, err := authenticateRequest(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, APIResponse{Success: false, Error: err.Error()})
		return
	}

	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Success: false, Error: "Method not allowed"})
		return
	}

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		writeJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Error: "Failed to read directory"})
		return
	}

	entries, err := os.ReadDir(uploadDir)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIResponse{Success: false, Error: "Failed to read files"})
		return
	}

	var files []map[string]interface{}
	for _, entry := range entries {
		if !entry.IsDir() {
			info, _ := entry.Info()
			files = append(files, map[string]interface{}{
				"name": entry.Name(),
				"size": info.Size(),
			})
		}
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    files,
		Message: fmt.Sprintf("%d files", len(files)),
	})
}

func HandleDownload(w http.ResponseWriter, r *http.Request) {
	_, err := authenticateRequest(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, APIResponse{Success: false, Error: err.Error()})
		return
	}

	filename := strings.TrimPrefix(r.URL.Path, "/api/data/download/")
	if filename == "" || strings.Contains(filename, "..") {
		http.NotFound(w, r)
		return
	}

	filepath := uploadDir + "/" + filename
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	http.ServeFile(w, r, filepath)
}
