package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/varunbhogayta-v11a/datautils/pkg/auth"
	"github.com/varunbhogayta-v11a/datautils/pkg/data"
	"github.com/varunbhogayta-v11a/datautils/pkg/db"
	"github.com/varunbhogayta-v11a/datautils/pkg/models"
	"github.com/varunbhogayta-v11a/datautils/pkg/operations"
	"github.com/spf13/cobra"
)

var (
	serverPort string
	serverHost string
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start HTTP API server with Swagger UI",
	Long: `Start the datautil HTTP API server with interactive Swagger documentation.
	
The server provides RESTful API endpoints for all data operations and includes
interactive Swagger UI at /swagger for testing and exploring the API.

Examples:
  datautil server
  datautil server --port 8080 --host 0.0.0.0`,
	Run: func(cmd *cobra.Command, args []string) {
		ensureDB()
		startServer()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVarP(&serverPort, "port", "p", "8080", "Port to listen on")
	serverCmd.Flags().StringVar(&serverHost, "host", "0.0.0.0", "Host to bind to")
}

func startServer() {
	mux := http.NewServeMux()

	mux.HandleFunc("/swagger.yaml", serveSwaggerSpec)
	mux.HandleFunc("/swagger/", serveSwaggerUI)
	mux.HandleFunc("/swagger", serveSwaggerUI)
	mux.HandleFunc("/api/health", handleHealth)

	mux.HandleFunc("/api/auth/register", handleAPIRegister)
	mux.HandleFunc("/api/auth/login", handleAPILogin)

	mux.HandleFunc("/api/data/filter", handleAPIFilter)
	mux.HandleFunc("/api/data/transform", handleAPITransform)
	mux.HandleFunc("/api/data/validate", handleAPIValidate)
	mux.HandleFunc("/api/data/export", handleAPIExport)
	mux.HandleFunc("/api/data/import", handleAPIImport)
	mux.HandleFunc("/api/data/upload", handleUpload)
	mux.HandleFunc("/api/data/files", handleListFiles)
	mux.HandleFunc("/api/data/download/", handleDownload)

	mux.HandleFunc("/api/db/query", handleAPIQuery)
	mux.HandleFunc("/api/db/insert", handleAPIInsert)
	mux.HandleFunc("/api/db/update", handleAPIUpdate)
	mux.HandleFunc("/api/db/delete", handleAPIDelete)

	mux.HandleFunc("/api/users", handleAPIUsers)
	mux.HandleFunc("/api/logs", handleAPILogs)

	mux.HandleFunc("/", handleRoot)

	addr := serverHost + ":" + serverPort
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	localAddr := "http://localhost:" + serverPort
	networkAddr := "http://127.0.0.1:" + serverPort

	ifaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range ifaces {
			if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
				continue
			}
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok {
					ip := ipnet.IP.To4()
					if ip != nil && (ip[0] == 10 || ip[0] == 192 || ip[0] == 172) {
						networkAddr = fmt.Sprintf("http://%s:%s", ipnet.IP.String(), serverPort)
						break
					}
				}
			}
		}
	}

	go func() {
		fmt.Printf("🚀 DataUtil API Server starting...\n")
		fmt.Printf("   Local:   %s\n", localAddr)
		fmt.Printf("   Network: %s\n", networkAddr)
		fmt.Printf("   Swagger: %s/swagger\n", localAddr)
		fmt.Printf("\n  VITE_like  press + ctrl + c to stop\n\n")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Server forced to shutdown: %v\n", err)
	}
	fmt.Println("Server stopped")
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html>
<head><title>DataUtil API</title></head>
<body>
<h1>DataUtil API Server</h1>
<p><a href="/swagger">Swagger Documentation</a></p>
<p><a href="/api/health">Health Check</a></p>
</body>
</html>`)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, resp APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

func getTokenFromRequest(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return r.URL.Query().Get("token")
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) == 2 && parts[0] == "Bearer" {
		return parts[1]
	}
	return ""
}

func authenticateRequest(r *http.Request) (*auth.Claims, error) {
	token := getTokenFromRequest(r)
	if token == "" {
		return nil, fmt.Errorf("authentication required")
	}
	jwt := auth.NewJWT()
	return jwt.ValidateToken(token)
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func handleAPIRegister(w http.ResponseWriter, r *http.Request) {
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

	var req RegisterRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "Invalid JSON"})
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "username, email, password required"})
		return
	}

	user, err := auth.Register(req.Username, req.Email, req.Password)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
		Message: "User registered successfully",
	})
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func handleAPILogin(w http.ResponseWriter, r *http.Request) {
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

	var req LoginRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "Invalid JSON"})
		return
	}

	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{Success: false, Error: "email and password required"})
		return
	}

	user, token, err := auth.Login(req.Email, req.Password)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, APIResponse{Success: false, Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"user": map[string]interface{}{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
				"role":     user.Role,
			},
			"token": token,
		},
		Message: "Login successful",
	})
}

type FilterRequest struct {
	Input  string `json:"input"`
	Where  string `json:"where,omitempty"`
	Select string `json:"select,omitempty"`
	Invert bool   `json:"invert,omitempty"`
	Output string `json:"output,omitempty"`
	Format string `json:"format,omitempty"`
}

func handleAPIFilter(w http.ResponseWriter, r *http.Request) {
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

func handleAPITransform(w http.ResponseWriter, r *http.Request) {
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

func handleAPIValidate(w http.ResponseWriter, r *http.Request) {
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

func handleAPIExport(w http.ResponseWriter, r *http.Request) {
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

func handleAPIImport(w http.ResponseWriter, r *http.Request) {
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

	colTypes := determineColumnTypesFromDataset(ds)

	if req.CreateTable {
		if err := createTableFromDataset(req.Table, ds.Headers, colTypes); err != nil {
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

const uploadDir = "./data/uploads"

func handleUpload(w http.ResponseWriter, r *http.Request) {
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

func handleListFiles(w http.ResponseWriter, r *http.Request) {
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

func handleDownload(w http.ResponseWriter, r *http.Request) {
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

type QueryRequest struct {
	SQL   string `json:"sql"`
	Limit int    `json:"limit,omitempty"`
}

func handleAPIQuery(w http.ResponseWriter, r *http.Request) {
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

func handleAPIInsert(w http.ResponseWriter, r *http.Request) {
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

func handleAPIUpdate(w http.ResponseWriter, r *http.Request) {
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

func handleAPIDelete(w http.ResponseWriter, r *http.Request) {
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

func handleAPIUsers(w http.ResponseWriter, r *http.Request) {
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

func handleAPILogs(w http.ResponseWriter, r *http.Request) {
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
			"created_at":  l.CreatedAt.Format(time.RFC3339),
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

func determineColumnTypesFromDataset(ds *data.Dataset) []string {
	types := make([]string, len(ds.Headers))

	for colIdx := range ds.Headers {
		for _, row := range ds.Rows {
			if colIdx >= len(row) {
				continue
			}
			val := row[colIdx]
			if val == "" {
				continue
			}

			if isIntegerAPI(val) {
				types[colIdx] = "INTEGER"
				break
			}
			if isFloatAPI(val) {
				types[colIdx] = "FLOAT"
				break
			}
		}
		if types[colIdx] == "" {
			types[colIdx] = "TEXT"
		}
	}

	return types
}

func isIntegerAPI(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			if c != '-' && c != '+' {
				return false
			}
		}
	}
	return true
}

func isFloatAPI(s string) bool {
	dotFound := false
	for _, c := range s {
		if c == '.' {
			if dotFound {
				return false
			}
			dotFound = true
		} else if c < '0' || c > '9' {
			if c != '-' && c != '+' {
				return false
			}
		}
	}
	return dotFound
}

func sanitizeColumnNameAPI(name string) string {
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

func convertToPostgresTypeAPI(sqlType string) string {
	switch sqlType {
	case "INTEGER":
		return "INTEGER"
	case "FLOAT":
		return "REAL"
	default:
		return "TEXT"
	}
}

func createTableFromDatasetAPI(tableName string, headers []string, types []string) error {
	driver := db.GetDriver()

	var sql string
	switch driver {
	case "sqlite":
		sql = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INTEGER PRIMARY KEY AUTOINCREMENT", tableName)
		for i, col := range headers {
			safeCol := sanitizeColumnNameAPI(col)
			sql += fmt.Sprintf(", %s %s", safeCol, types[i])
		}
		sql += ")"
	case "postgres":
		sql = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id SERIAL PRIMARY KEY", tableName)
		for i, col := range headers {
			safeCol := sanitizeColumnNameAPI(col)
			pgType := convertToPostgresTypeAPI(types[i])
			sql += fmt.Sprintf(", %s %s", safeCol, pgType)
		}
		sql += ")"
	default:
		sql = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INTEGER PRIMARY KEY", tableName)
		for i, col := range headers {
			sql += fmt.Sprintf(", %s %s", col, types[i])
		}
		sql += ")"
	}

	_, err := db.DB.Exec(sql)
	return err
}

func serveSwaggerSpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	w.Header().Set("Content-Disposition", "inline")
	w.Write([]byte(swaggerSpec))
}

func serveSwaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(swaggerHTML))
}

var swaggerSpec = `openapi: 3.0.3
info:
  title: DataUtil API
  description: |
    RESTful API for dataset processing, filtering, transformation, validation, and database operations.
    Supports CSV, JSON, XML, and Excel formats with JWT-based authentication.
  version: 1.0.0
  contact:
    name: API Support
    email: support@datautil.io

servers:
  - url: http://localhost:8080/api
    description: Local development server

security:
  - bearerAuth: []

tags:
  - name: Authentication
    description: User registration and login endpoints
  - name: Data Operations
    description: Filter, transform, validate, and export datasets
  - name: Database
    description: Query and manipulate database tables
  - name: Users
    description: User management and logs

paths:
  /health:
    get:
      summary: Health check endpoint
      tags:
        - Health
      responses:
        '200':
          description: Server is healthy
          content:
            application/json:
              schema:
                type: object
                properties:
                  status:
                    type: string
                  timestamp:
                    type: string

  /auth/register:
    post:
      summary: Register a new user
      tags:
        - Authentication
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - username
                - email
                - password
              properties:
                username:
                  type: string
                  description: Unique username
                email:
                  type: string
                  format: email
                  description: User email address
                password:
                  type: string
                  format: password
                  description: User password
      responses:
        '201':
          description: User created successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/RegisterResponse'
        '400':
          description: Invalid request or user already exists

  /auth/login:
    post:
      summary: Login and get JWT token
      tags:
        - Authentication
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - email
                - password
              properties:
                email:
                  type: string
                  format: email
                password:
                  type: string
                  format: password
      responses:
        '200':
          description: Login successful
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/LoginResponse'
        '401':
          description: Invalid credentials

  /data/filter:
    post:
      summary: Filter rows from dataset
      tags:
        - Data Operations
      security:
        - bearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - input
              properties:
                input:
                  type: string
                  description: Input file path (required)
                where:
                  type: string
                  description: Filter condition (e.g., "age > 25")
                select:
                  type: string
                  description: Columns to select (comma-separated)
                invert:
                  type: boolean
                  description: Invert filter (exclude matching rows)
                output:
                  type: string
                  description: Output file path
                format:
                  type: string
                  enum: [csv, json, xml]
      responses:
        '200':
          description: Filtered dataset
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DataResponse'
        '401':
          description: Authentication required

  /data/transform:
    post:
      summary: Transform dataset (add, remove, rename columns)
      tags:
        - Data Operations
      security:
        - bearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - input
              properties:
                input:
                  type: string
                add:
                  type: string
                  description: 'Add column (format: "name=expression")'
                remove:
                  type: string
                  description: Columns to remove (comma-separated)
                rename:
                  type: string
                  description: 'Rename column (format: "old:new")'
                output:
                  type: string
                format:
                  type: string
                  enum: [csv, json, xml]
      responses:
        '200':
          description: Transformed dataset
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DataResponse'

  /data/validate:
    post:
      summary: Validate dataset against schema
      tags:
        - Data Operations
      security:
        - bearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - input
              properties:
                input:
                  type: string
                required:
                  type: string
                  description: Required columns (comma-separated)
                types:
                  type: string
                  description: 'Column types (format: "col:type")'
      responses:
        '200':
          description: Validation result
          content:
            application/json:
              schema:
                type: object
                properties:
                  success:
                    type: boolean
                  data:
                    type: object
                    properties:
                      valid:
                        type: boolean
                      errors:
                        type: array
                        items:
                          type: string
                      rows:
                        type: integer

  /data/export:
    post:
      summary: Export dataset to different format
      tags:
        - Data Operations
      security:
        - bearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - input
                - to
              properties:
                input:
                  type: string
                output:
                  type: string
                to:
                  type: string
                  enum: [csv, json, xml]
                pretty:
                  type: boolean
                stats:
                  type: boolean
      responses:
        '200':
          description: Exported data
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DataResponse'

  /data/import:
    post:
      summary: Import file data to database
      tags:
        - Data Operations
      security:
        - bearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - source
                - table
              properties:
                source:
                  type: string
                  description: Source file path
                table:
                  type: string
                  description: Target database table
                create:
                  type: boolean
                ifNotExists:
                  type: boolean
                truncate:
                  type: boolean
      responses:
        '200':
          description: Import successful
          content:
            application/json:
              schema:
                type: object
                properties:
                  success:
                    type: boolean
                  data:
                    type: object
                    properties:
                      table:
                        type: string
                      rows:
                        type: integer

  /db/query:
    post:
      summary: Execute SQL SELECT query
      tags:
        - Database
      security:
        - bearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - sql
              properties:
                sql:
                  type: string
                  description: SQL SELECT query
                limit:
                  type: integer
                  default: 100
      responses:
        '200':
          description: Query results
          content:
            application/json:
              schema:
                type: object
                properties:
                  success:
                    type: boolean
                  data:
                    type: object
                    properties:
                      columns:
                        type: array
                        items:
                          type: string
                      rows:
                        type: array
                        items:
                          type: array
                          items:
                            type: string
                      count:
                        type: integer

  /db/insert:
    post:
      summary: Insert row into database table
      tags:
        - Database
      security:
        - bearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - table
                - values
              properties:
                table:
                  type: string
                values:
                  type: string
                  description: Values in format "col1=val1,col2=val2"
      responses:
        '201':
          description: Insert successful

  /db/update:
    post:
      summary: Update rows in database table
      tags:
        - Database
      security:
        - bearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - table
                - set
              properties:
                table:
                  type: string
                set:
                  type: string
                  description: SET clause (e.g., "age=30")
                where:
                  type: string
                  description: WHERE clause
      responses:
        '200':
          description: Update successful

  /db/delete:
    post:
      summary: Delete rows from database table
      tags:
        - Database
      security:
        - bearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - table
              properties:
                table:
                  type: string
                where:
                  type: string
                  description: WHERE clause
      responses:
        '200':
          description: Delete successful

  /users:
    get:
      summary: List all users
      tags:
        - Users
      security:
        - bearerAuth: []
      responses:
        '200':
          description: User list
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: object
                    properties:
                      users:
                        type: array
                        items:
                          type: object

  /logs:
    get:
      summary: Get operation logs
      tags:
        - Users
      security:
        - bearerAuth: []
      responses:
        '200':
          description: Operation logs
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: object

components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

  schemas:
    RegisterResponse:
      type: object
      properties:
        success:
          type: boolean
        data:
          type: object
          properties:
            id:
              type: integer
            username:
              type: string
            email:
              type: string
            role:
              type: string
        message:
          type: string

    LoginResponse:
      type: object
      properties:
        success:
          type: boolean
        data:
          type: object
          properties:
            user:
              type: object
            token:
              type: string
        message:
          type: string

    DataResponse:
      type: object
      properties:
        success:
          type: boolean
        data:
          type: object
          properties:
            headers:
              type: array
              items:
                type: string
            rows:
              type: array
              items:
                type: array
                items:
                  type: string
            count:
              type: integer
        message:
          type: string

    HealthResponse:
      type: object
      properties:
        status:
          type: string
        timestamp:
          type: string

    ValidationResult:
      type: object
      properties:
        success:
          type: boolean
        data:
          type: object
          properties:
            valid:
              type: boolean
            errors:
              type: array
              items:
                type: string
            rows:
              type: integer
        message:
          type: string

    ImportResult:
      type: object
      properties:
        success:
          type: boolean
        data:
          type: object
          properties:
            table:
              type: string
            rows:
              type: integer
        message:
          type: string

    QueryResult:
      type: object
      properties:
        success:
          type: boolean
        data:
          type: object
          properties:
            columns:
              type: array
              items:
                type: string
            rows:
              type: array
              items:
                type: array
            count:
              type: integer
        message:
          type: string

    UserList:
      type: object
      properties:
        success:
          type: boolean
        data:
          type: object
          properties:
            users:
              type: array
              items:
                type: object
                properties:
                  id:
                    type: integer
                  username:
                    type: string
                  email:
                    type: string
                  role:
                    type: string
                  active:
                    type: boolean
        message:
          type: string

    LogsResponse:
      type: object
      properties:
        success:
          type: boolean
        data:
          type: object
          properties:
            logs:
              type: array
              items:
                type: object
                properties:
                  id:
                    type: integer
                  userId:
                    type: integer
                  operation:
                    type: string
                  inputFile:
                    type: string
                  outputFile:
                    type: string
                  details:
                    type: string
                  createdAt:
                    type: string
        message:
          type: string

    ErrorResponse:
      type: object
      properties:
        success:
          type: boolean
        error:
          type: string
        message:
          type: string
`

var swaggerHTML = `<!DOCTYPE html>
<html>
<head>
  <title>DataUtil API - Swagger UI</title>
  <link rel="stylesheet" type="text/css" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.10.5/swagger-ui.css" />
  <style>
    body { margin: 0; padding: 0; }
    .swagger-ui .topbar { display: none; }
    .swagger-ui .info .title { font-size: 2.5em; }
  </style>
</head>
<body>
	<div id="swagger-ui"></div>
	<div class="loading" id="loading">Loading Swagger UI...</div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.10.5/swagger-ui-bundle.js" charset="UTF-8"></script>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.10.5/swagger-ui-standalone-preset.js" charset="UTF-8"></script>
	<script>
		window.onload = function() {
			var loading = document.getElementById("loading");
			try {
				if (window.SwaggerUIBundle) {
					var opts = {
												url: "/swagger.yaml",
						dom_id: "#swagger-ui",
						deepLinking: true,
						docExpansion: "list",
						filter: true
					};
					if (SwaggerUIBundle.presets && SwaggerUIBundle.presets.apis && SwaggerUIBundle.standalonePreset) {
						opts.presets = [SwaggerUIBundle.presets.apis, SwaggerUIBundle.standalonePreset];
					}
					if (SwaggerUIBundle.plugins && SwaggerUIBundle.plugins.DownloadUrl) {
						opts.plugins = [SwaggerUIBundle.plugins.DownloadUrl];
					}
					window.ui = SwaggerUIBundle(opts);
				} else if (window.SwaggerUI) {
					var opts2 = {
						url: "/swagger.yaml",
						dom_id: "#swagger-ui",
						deepLinking: true,
						docExpansion: "list",
						filter: true
					};
					if (SwaggerUI.presets && SwaggerUI.presets.apis && SwaggerUI.standalonePreset) {
						opts2.presets = [SwaggerUI.presets.apis, SwaggerUI.standalonePreset];
					}
					window.ui = SwaggerUI(opts2);
				} else {
					throw new Error('Swagger UI bundle not found');
				}

				if (loading) loading.style.display = "none";
			} catch (e) {
				if (loading) loading.innerHTML = "Error loading Swagger UI: " + e.message;
				else console.error('Error loading Swagger UI:', e);
			}
		};
	</script>
</body>
</html>`
