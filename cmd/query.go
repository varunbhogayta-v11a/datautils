package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/improwised/datautil/pkg/auth"
	"github.com/improwised/datautil/pkg/db"
	"github.com/improwised/datautil/pkg/models"
	"github.com/spf13/cobra"
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Execute SQL query on database",
	Long: `Execute SELECT queries on database tables.

Examples:
  datautil query --sql "SELECT * FROM users"
  datautil query --sql "SELECT * FROM test_data WHERE age > 25" --token "YOUR_TOKEN"`,
	Run: func(cmd *cobra.Command, args []string) {
		requireAuth(cmd)
		ensureDB()

		tempUser := &models.User{Role: currentUser.Role}
		if !tempUser.HasPermission("read") {
			fmt.Fprintln(os.Stderr, "Error: Permission denied")
			os.Exit(1)
		}

		sqlQuery = strings.TrimSpace(sqlQuery)
		if sqlQuery == "" {
			fmt.Fprintln(os.Stderr, "Error: --sql required")
			os.Exit(1)
		}

		upperQuery := strings.ToUpper(sqlQuery)
		if !strings.HasPrefix(upperQuery, "SELECT") && !strings.HasPrefix(upperQuery, "PRAGMA") && !strings.HasPrefix(upperQuery, "SHOW") {
			fmt.Fprintln(os.Stderr, "Error: Only SELECT queries allowed")
			os.Exit(1)
		}

		rows, err := db.DB.Query(sqlQuery)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error executing query: %v\n", err)
			os.Exit(1)
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting columns: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(strings.Join(columns, "\t"))

		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		rowCount := 0
		for rows.Next() {
			err := rows.Scan(valuePtrs...)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error scanning row: %v\n", err)
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
			fmt.Println(strings.Join(rowValues, "\t"))
			rowCount++
		}

		fmt.Printf("\n(%d rows)\n", rowCount)
		auth.LogOperation(currentUser.UserID, "query", sqlQuery, "", "")
	},
}

var (
	sqlQuery string
	limit    int
)

func init() {
	rootCmd.AddCommand(queryCmd)
	queryCmd.Flags().StringVarP(&sqlQuery, "sql", "s", "", "SQL query to execute (required)")
	queryCmd.Flags().IntVarP(&limit, "limit", "l", 100, "Limit number of results")
}
