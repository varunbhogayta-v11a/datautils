package cmd

import (
	"fmt"
	"os"

	"github.com/varunbhogayta-v11a/datautils/pkg/auth"
	"github.com/varunbhogayta-v11a/datautils/pkg/db"
	"github.com/varunbhogayta-v11a/datautils/pkg/models"
	"github.com/varunbhogayta-v11a/datautils/pkg/repo"
	"github.com/spf13/cobra"
)

var dbInitCmd = &cobra.Command{
	Use:   "init-db",
	Short: "Initialize database and run migrations",
	Long:  `Connect to database and create tables`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := db.GetConfig()

		fmt.Println("Database driver:", cfg.Driver)
		fmt.Println("Database name:", cfg.DBName)

		fmt.Println("Creating database if not exists...")
		if err := db.CreateDatabase(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Warning creating database: %v\n", err)
		}

		fmt.Println("Connecting to database...")
		if err := db.Connect(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Running migrations...")
		tables := models.GetTableName(cfg.Driver)
		for tableName, createSQL := range tables {
			fmt.Printf("Creating table: %s\n", tableName)
			if _, err := db.DB.Exec(createSQL); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating %s: %v\n", tableName, err)
				continue
			}
			fmt.Printf("✓ Table %s created\n", tableName)
		}

		fmt.Println("✓ Database initialized successfully")
	},
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new user",
	Long:  `Create a new user account`,
	Run: func(cmd *cobra.Command, args []string) {
		ensureDB()

		user, err := auth.Register(username, email, password)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ User registered successfully: %s (%s)\n", user.Username, user.Email)
	},
}

var (
	username string
	email    string
	password string
)

func init() {
	rootCmd.AddCommand(dbInitCmd)
	rootCmd.AddCommand(registerCmd)

	registerCmd.Flags().StringVarP(&username, "username", "u", "", "Username (required)")
	registerCmd.Flags().StringVarP(&email, "email", "e", "", "Email (required)")
	registerCmd.Flags().StringVarP(&password, "password", "p", "", "Password (required)")

	registerCmd.MarkFlagRequired("username")
	registerCmd.MarkFlagRequired("email")
	registerCmd.MarkFlagRequired("password")
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login and get JWT token",
	Long:  `Authenticate and receive a JWT token`,
	Run: func(cmd *cobra.Command, args []string) {
		ensureDB()

		user, token, err := auth.Login(email, password)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Login successful\n")
		fmt.Printf("User: %s | Role: %s\n", user.Username, user.Role)
		fmt.Printf("\nJWT Token:\n%s\n", token)
		fmt.Printf("\nUse: datautil --token \"%s\" <command>\n", token)
	},
}

var token string
var currentUser *auth.Claims

func requireAuth(cmd *cobra.Command) {
	if token == "" {
		fmt.Fprintln(os.Stderr, "Error: Authentication required. Use --token or run 'datautil login'")
		os.Exit(1)
	}

	jwt := auth.NewJWT()
	claims, err := jwt.ValidateToken(token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Invalid token: %v\n", err)
		os.Exit(1)
	}
	currentUser = claims
}

func ensureDB() {
	if db.DB == nil {
		cfg := db.GetConfig()
		if err := db.Connect(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
			os.Exit(1)
		}
	}
	auth.SetUserRepository(repo.NewUserRepository())
	auth.SetOperationLogRepository(repo.NewOperationLogRepository())
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringVarP(&email, "email", "e", "", "Email (required)")
	loginCmd.Flags().StringVarP(&password, "password", "p", "", "Password (required)")
	loginCmd.MarkFlagRequired("email")
	loginCmd.MarkFlagRequired("password")

	rootCmd.PersistentFlags().StringVar(&token, "token", "", "JWT token for authentication")
}
