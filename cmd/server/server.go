package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/varunbhogayta-v11a/datautils/cmd/server/handlers"
	"github.com/varunbhogayta-v11a/datautils/cmd/server/templates"
	"github.com/varunbhogayta-v11a/datautils/pkg/db"
)

var serverPort string
var serverHost string

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
	serverCmd.Flags().StringVarP(&serverPort, "port", "p", "8080", "Port to listen on")
	serverCmd.Flags().StringVar(&serverHost, "host", "0.0.0.0", "Host to bind to")
}

func ensureDB() {
	if err := db.ConnectDefault(); err != nil {
		fmt.Fprintf(os.Stderr, "Database connection failed: %v\n", err)
		os.Exit(1)
	}
}

func startServer() {
	mux := http.NewServeMux()
	registerRoutes(mux)

	addr := serverHost + ":" + serverPort
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	printServerInfo(server)

	go func() {
		fmt.Printf("🚀 DataUtil API Server starting...")
		fmt.Printf("\n press + ctrl + c to stop\n\n")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	}()

	waitForShutdown(server)
}

func registerRoutes(mux *http.ServeMux) {
	routes := []struct {
		path    string
		handler http.HandlerFunc
	}{
		{"/swagger.yaml", templates.ServeSwaggerSpec},
		{"/swagger.json", templates.ServeSwaggerSpec},
		{"/swagger", templates.ServeSwaggerUI},
		{"/swagger/", templates.ServeSwaggerUI},
		{"/api/health", handlers.HandleHealth},
		{"/api/auth/register", handlers.HandleAPIRegister},
		{"/api/auth/login", handlers.HandleAPILogin},
		{"/api/data/filter", handlers.HandleAPIFilter},
		{"/api/data/transform", handlers.HandleAPITransform},
		{"/api/data/validate", handlers.HandleAPIValidate},
		{"/api/data/export", handlers.HandleAPIExport},
		{"/api/data/import", handlers.HandleAPIImport},
		{"/api/data/upload", handlers.HandleUpload},
		{"/api/data/files", handlers.HandleListFiles},
		{"/api/data/download/", handlers.HandleDownload},
		{"/api/db/query", handlers.HandleAPIQuery},
		{"/api/db/insert", handlers.HandleAPIInsert},
		{"/api/db/update", handlers.HandleAPIUpdate},
		{"/api/db/delete", handlers.HandleAPIDelete},
		{"/api/users", handlers.HandleAPIUsers},
		{"/api/logs", handlers.HandleAPILogs},
		{"/", handlers.HandleRoot},
	}

	for _, r := range routes {
		mux.HandleFunc(r.path, r.handler)
	}
}

func printServerInfo(server *http.Server) {
	localAddr := "http://localhost:" + serverPort
	networkAddr := getNetworkAddr()

	fmt.Printf("🚀 DataUtil API Server starting...\n")
	fmt.Printf("   Local:   %s\n", localAddr)
	fmt.Printf("   Network: %s\n", networkAddr)
	fmt.Printf("   Swagger: %s/swagger\n", localAddr)
}

func getNetworkAddr() string {
	networkAddr := "http://127.0.0.1:" + serverPort

	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				ip := ipnet.IP.To4()
				if ip != nil && (ip[0] == 10 || ip[0] == 192 || ip[0] == 172) {
					return fmt.Sprintf("http://%s:%s", ipnet.IP.String(), serverPort)
				}
			}
		}
	}
	return networkAddr
}

func waitForShutdown(server *http.Server) {
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

func AddServerCommand(root *cobra.Command) {
	root.AddCommand(serverCmd)
}
