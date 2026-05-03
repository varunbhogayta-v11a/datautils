package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/varunbhogayta-v11a/datautils/pkg/auth"
	"github.com/varunbhogayta-v11a/datautils/pkg/data"
	"github.com/varunbhogayta-v11a/datautils/pkg/db"
	"github.com/varunbhogayta-v11a/datautils/pkg/operations"
	"github.com/spf13/cobra"
)

var (
	requiredCols string
	colTypes     string
	pattern      string
	strictMode   bool
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate dataset against schema",
	Long: `Validate data against defined schema or rules.

Examples:
  datautil validate --input data.csv --required name,email,age --token "YOUR_JWT"`,
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

		var required []string
		if requiredCols != "" {
			required = strings.Split(requiredCols, ",")
		}

		typesMap := make(map[string]string)
		if colTypes != "" {
			for _, t := range strings.Split(colTypes, ",") {
				parts := strings.SplitN(t, ":", 2)
				if len(parts) == 2 {
					typesMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
			}
		}

		result := operations.ValidateDataset(ds, required, typesMap)
		_ = result

		if result.Valid {
			fmt.Println("✓ Validation passed")
			fmt.Printf("  Rows: %d, Columns: %d\n", ds.RowCount(), ds.ColCount())
			auth.LogOperation(currentUser.UserID, "validate", inputFile, "", requiredCols+colTypes)
		} else {
			fmt.Println("✗ Validation failed:")
			for _, err := range result.Errors {
				fmt.Printf("  - %s\n", err)
			}
			if !strictMode {
				fmt.Println("\nNote: Run with --strict to fail on validation errors")
			} else {
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input file (required)")
	validateCmd.Flags().StringVarP(&requiredCols, "required", "r", "", "Required columns (comma-separated)")
	validateCmd.Flags().StringVarP(&colTypes, "types", "t", "", "Column types (format: col:type, comma-separated)")
	validateCmd.Flags().StringVarP(&pattern, "pattern", "p", "", "Pattern for validation (regex)")
	validateCmd.Flags().BoolVarP(&strictMode, "strict", "", false, "Exit with error on validation failure")
	if err := validateCmd.MarkFlagRequired("input"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
