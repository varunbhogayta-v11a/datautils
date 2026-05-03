package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/varunbhogayta-v11a/datautils/pkg/auth"
	"github.com/varunbhogayta-v11a/datautils/pkg/data"
	"github.com/varunbhogayta-v11a/datautils/pkg/db"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import CSV data to database",
	Long: `Import CSV file data into a database table.

Examples:
  datautil import --input data.csv --table users
  datautil import --input data.csv --table products --if-not-exists
  datautil import --input data.csv --table users --truncate`,
	Run: func(cmd *cobra.Command, args []string) {
		requireAuth(cmd)

		if db.DB == nil {
			cfg := db.GetConfig()
			if err := db.Connect(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
				os.Exit(1)
			}
		}

		ds, err := data.ReadFile(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}

		if ds.RowCount() == 0 {
			fmt.Println("No data to import")
			os.Exit(0)
		}

		fmt.Printf("Importing %d rows with %d columns...\n", ds.RowCount(), ds.ColCount())

		// Determine SQL types based on data
		colTypes := determineColumnTypes(ds)

		// Create table if needed
		if createTable {
			if err := createTableFromDataset(tableName, ds.Headers, colTypes); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating table: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✓ Created table: %s\n", tableName)
		}

		// Check if table exists
		if ifNotExists {
			var count int
			switch db.GetDriver() {
			case "sqlite":
				err := db.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='%s'", tableName)).Scan(&count)
				if err == nil && count > 0 {
					fmt.Printf("Table %s already exists, skipping import\n", tableName)
					return
				}
			case "mysql":
				err := db.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='%s' AND table_name='%s'",
					db.GetConfig().DBName, tableName)).Scan(&count)
				if err == nil && count > 0 {
					fmt.Printf("Table %s already exists, skipping import\n", tableName)
					return
				}
			case "postgres":
				err := db.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name='%s'",
					tableName)).Scan(&count)
				if err == nil && count > 0 {
					fmt.Printf("Table %s already exists, skipping import\n", tableName)
					return
				}
			}
		}

		// Truncate if requested
		if truncateTable {
			_, err := db.DB.Exec(fmt.Sprintf("DELETE FROM %s", tableName))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error truncating table: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✓ Truncated table: %s\n", tableName)
		}

		// Insert data in batches
		batchSize := 1000
		totalInserted := 0

		for i := 0; i < len(ds.Rows); i += batchSize {
			end := i + batchSize
			if end > len(ds.Rows) {
				end = len(ds.Rows)
			}

			// Build insert statement
			columns := strings.Join(ds.Headers, ", ")
			placeholders := make([]string, len(ds.Headers))
			for j := range ds.Headers {
				placeholders[j] = "?"
			}

			query := fmt.Sprintf("INSERT INTO %s (%s) VALUES ", tableName, columns)
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
				fmt.Fprintf(os.Stderr, "Error inserting batch: %v\n", err)
				os.Exit(1)
			}

			totalInserted += end - i
			fmt.Printf("Inserted %d/%d rows\n", totalInserted, len(ds.Rows))
		}

		fmt.Printf("✓ Successfully imported %d rows into table: %s\n", totalInserted, tableName)

		// Log operation
		if currentUser != nil {
			auth.LogOperation(currentUser.UserID, "import", inputFile, tableName, fmt.Sprintf("%d rows", totalInserted))
		}
	},
}

var (
	tableName     string
	createTable   bool
	ifNotExists   bool
	truncateTable bool
)

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input CSV file (required)")
	importCmd.Flags().StringVarP(&tableName, "table", "t", "", "Table name (required)")
	importCmd.Flags().BoolVarP(&createTable, "create", "c", false, "Create table if not exists")
	importCmd.Flags().BoolVar(&ifNotExists, "if-not-exists", false, "Skip if table exists")
	importCmd.Flags().BoolVar(&truncateTable, "truncate", false, "Truncate table before import")

	importCmd.MarkFlagRequired("input")
	importCmd.MarkFlagRequired("table")
}

func determineColumnTypes(ds *data.Dataset) []string {
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

			// Check if integer
			if isInteger(val) {
				types[colIdx] = "INTEGER"
				break
			}
			// Check if float
			if isFloat(val) {
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

func isInteger(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			if c != '-' && c != '+' {
				return false
			}
		}
	}
	return true
}

func isFloat(s string) bool {
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

func createTableFromDataset(tableName string, headers []string, types []string) error {
	driver := db.GetDriver()

	var sql string
	switch driver {
	case "sqlite":
		sql = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INTEGER PRIMARY KEY AUTOINCREMENT", tableName)
		for i, col := range headers {
			safeCol := sanitizeColumnName(col)
			sql += fmt.Sprintf(", %s %s", safeCol, types[i])
		}
		sql += ")"
	case "mysql":
		sql = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INT AUTO_INCREMENT PRIMARY KEY", tableName)
		for i, col := range headers {
			safeCol := sanitizeColumnName(col)
			sql += fmt.Sprintf(", %s %s", safeCol, types[i])
		}
		sql += ") ENGINE=InnoDB"
	case "postgres":
		sql = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id SERIAL PRIMARY KEY", tableName)
		for i, col := range headers {
			safeCol := sanitizeColumnName(col)
			pgType := convertToPostgresType(types[i])
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

func sanitizeColumnName(name string) string {
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	// Remove invalid characters
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

func convertToPostgresType(sqlType string) string {
	switch sqlType {
	case "INTEGER":
		return "INTEGER"
	case "FLOAT":
		return "REAL"
	default:
		return "TEXT"
	}
}

func init() {
}
