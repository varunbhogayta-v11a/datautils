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

var insertCmd = &cobra.Command{
	Use:   "insert",
	Short: "Insert data into database table",
	Long: `Insert rows into a database table.

Examples:
  datautil insert --table users --values "name=John,age=25,city=NYC" --token "YOUR_TOKEN"
  datautil insert --table products --values "name=Laptop,price=999.99" --token "YOUR_TOKEN"`,
	Run: func(cmd *cobra.Command, args []string) {
		requireAuth(cmd)
		ensureDB()

		tempUser := &models.User{Role: currentUser.Role}
		if !tempUser.HasPermission("write") {
			fmt.Fprintln(os.Stderr, "Error: Permission denied - write access required")
			os.Exit(1)
		}

		if tableName == "" || values == "" {
			fmt.Fprintln(os.Stderr, "Error: --table and --values are required")
			os.Exit(1)
		}

		// Parse values: "name=John,age=25,city=NYC"
		pairs := strings.Split(values, ",")
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
			tableName, strings.Join(cols, ", "), strings.Join(vals, ", "))

		result, err := db.DB.Exec(query, argsSlice...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error inserting data: %v\n", err)
			os.Exit(1)
		}

		rowsAffected, _ := result.RowsAffected()
		fmt.Printf("✓ Inserted %d row(s) into %s\n", rowsAffected, tableName)

		auth.LogOperation(currentUser.UserID, "insert", tableName, "", values)
	},
}

var values string

func init() {
	rootCmd.AddCommand(insertCmd)
	insertCmd.Flags().StringVarP(&tableName, "table", "t", "", "Table name (required)")
	insertCmd.Flags().StringVarP(&values, "values", "v", "", "Values to insert (format: col1=val1,col2=val2)")
	insertCmd.MarkFlagRequired("table")
	insertCmd.MarkFlagRequired("values")
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update data in database table",
	Long: `Update rows in a database table.

Examples:
  datautil update --table users --set "age=30" --where "name=John" --token "YOUR_TOKEN"
  datautil update --table products --set "price=899.99" --where "name=Laptop" --token "YOUR_TOKEN"`,
	Run: func(cmd *cobra.Command, args []string) {
		requireAuth(cmd)
		ensureDB()

		tempUser := &models.User{Role: currentUser.Role}
		if !tempUser.HasPermission("write") {
			fmt.Fprintln(os.Stderr, "Error: Permission denied - write access required")
			os.Exit(1)
		}

		if tableName == "" || setValues == "" {
			fmt.Fprintln(os.Stderr, "Error: --table and --set are required")
			os.Exit(1)
		}

		query := fmt.Sprintf("UPDATE %s SET %s", tableName, setValues)
		if whereCond != "" {
			query += " WHERE " + whereCond
		}

		result, err := db.DB.Exec(query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating data: %v\n", err)
			os.Exit(1)
		}

		rowsAffected, _ := result.RowsAffected()
		fmt.Printf("✓ Updated %d row(s) in %s\n", rowsAffected, tableName)

		auth.LogOperation(currentUser.UserID, "update", tableName, "", setValues+" "+whereCond)
	},
}

var (
	setValues   string
	deleteWhere string
)

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringVarP(&tableName, "table", "t", "", "Table name (required)")
	updateCmd.Flags().StringVarP(&setValues, "set", "s", "", "Set clause (format: col=val)")
	updateCmd.Flags().StringVarP(&whereCond, "where", "w", "", "Where clause")
	updateCmd.MarkFlagRequired("table")
	updateCmd.MarkFlagRequired("set")
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete data from database table",
	Long: `Delete rows from a database table.

Examples:
  datautil delete --table users --where "name=John" --token "YOUR_TOKEN"
  datautil delete --table products --where "price<10" --token "YOUR_TOKEN"`,
	Run: func(cmd *cobra.Command, args []string) {
		requireAuth(cmd)
		ensureDB()

		tempUser := &models.User{Role: currentUser.Role}
		if !tempUser.HasPermission("delete") {
			fmt.Fprintln(os.Stderr, "Error: Permission denied - delete access required")
			os.Exit(1)
		}

		if tableName == "" {
			fmt.Fprintln(os.Stderr, "Error: --table is required")
			os.Exit(1)
		}

		query := fmt.Sprintf("DELETE FROM %s", tableName)
		if deleteWhere != "" {
			query += " WHERE " + deleteWhere
		}

		result, err := db.DB.Exec(query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting data: %v\n", err)
			os.Exit(1)
		}

		rowsAffected, _ := result.RowsAffected()
		fmt.Printf("✓ Deleted %d row(s) from %s\n", rowsAffected, tableName)

		auth.LogOperation(currentUser.UserID, "delete", tableName, "", deleteWhere)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringVarP(&tableName, "table", "t", "", "Table name (required)")
	deleteCmd.Flags().StringVarP(&deleteWhere, "where", "w", "", "Where clause")
	deleteCmd.MarkFlagRequired("table")
}
