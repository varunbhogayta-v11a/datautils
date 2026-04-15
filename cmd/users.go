package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/improwised/datautil/pkg/db"
	"github.com/improwised/datautil/pkg/models"
	"github.com/spf13/cobra"
)

var usersListCmd = &cobra.Command{
	Use:   "users",
	Short: "List all users",
	Long:  `Display all registered users (admin only)`,
	Run: func(cmd *cobra.Command, args []string) {
		if token == "" {
			fmt.Fprintln(os.Stderr, "Error: Authentication required")
			os.Exit(1)
		}

		if db.DB == nil {
			cfg := db.GetConfig()
			if err := db.Connect(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "Error connecting: %v\n", err)
				os.Exit(1)
			}
		}

		rows, err := db.DB.Query("SELECT id, username, email, role, active FROM users")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer rows.Close()

		fmt.Println("Registered Users:")
		fmt.Println("ID\tUsername\tEmail\t\tRole\tActive")
		fmt.Println(strings.Repeat("-", 60))
		for rows.Next() {
			var u models.User
			if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.Active); err != nil {
				fmt.Fprintf(os.Stderr, "Error scanning: %v\n", err)
				continue
			}
			fmt.Printf("%d\t%s\t\t%s\t%s\t%v\n", u.ID, u.Username, u.Email, u.Role, u.Active)
		}
	},
}

func init() {
	rootCmd.AddCommand(usersListCmd)
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show operation logs",
	Long:  `Display operation history for the current user`,
	Run: func(cmd *cobra.Command, args []string) {
		if token == "" {
			fmt.Fprintln(os.Stderr, "Error: Authentication required")
			os.Exit(1)
		}

		if db.DB == nil {
			cfg := db.GetConfig()
			if err := db.Connect(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "Error connecting: %v\n", err)
				os.Exit(1)
			}
		}

		query := "SELECT id, user_id, operation, input_file, output_file, created_at FROM operation_logs ORDER BY created_at DESC LIMIT 50"

		rows, err := db.DB.Query(query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer rows.Close()

		fmt.Println("Operation History:")
		fmt.Println("ID\tOperation\tInput\t\tOutput\t\tTime")
		fmt.Println(strings.Repeat("-", 80))
		for rows.Next() {
			var l models.OperationLog
			if err := rows.Scan(&l.ID, &l.UserID, &l.Operation, &l.InputFile, &l.OutputFile, &l.CreatedAt); err != nil {
				fmt.Fprintf(os.Stderr, "Error scanning: %v\n", err)
				continue
			}
			fmt.Printf("%d\t%s\t\t%s\t\t%s\t%s\n", l.ID, l.Operation, l.InputFile, l.OutputFile, l.CreatedAt.Format("2006-01-02 15:04"))
		}
		if rows.Err() != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", rows.Err())
		}
		fmt.Println("Note: Use --token with admin role to see all logs")
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)
}
