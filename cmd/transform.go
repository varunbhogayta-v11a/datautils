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
	addCol     string
	removeCols string
	renameCol  string
	updateCol  string
)

var transformCmd = &cobra.Command{
	Use:   "transform",
	Short: "Transform dataset (add, remove, rename columns)",
	Long: `Transform data by adding, removing, or renaming columns.

Examples:
  datautil transform --input data.csv --add "full_name=first+last" --token "YOUR_JWT"`,
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

		if addCol != "" {
			parts := strings.SplitN(addCol, "=", 2)
			if len(parts) == 2 {
				ds = operations.AddColumn(ds, strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			}
		}

		if removeCols != "" {
			ds = operations.RemoveColumns(ds, strings.Split(removeCols, ","))
		}

		if renameCol != "" {
			parts := strings.SplitN(renameCol, ":", 2)
			if len(parts) == 2 {
				ds = operations.RenameColumn(ds, strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			}
		}

		if outputFile != "" {
			if err := data.WriteFile(ds, outputFile, format); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Transformed data saved to %s\n", outputFile)
			auth.LogOperation(currentUser.UserID, "transform", inputFile, outputFile, addCol+removeCols+renameCol)
		} else {
			data.PrintDataset(ds)
		}
	},
}

func init() {
	rootCmd.AddCommand(transformCmd)
	transformCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input file (required)")
	transformCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file")
	transformCmd.Flags().StringVarP(&format, "format", "f", "", "Output format (csv, json, xml)")
	transformCmd.Flags().StringVarP(&addCol, "add", "a", "", "Add column (format: name=expression)")
	transformCmd.Flags().StringVarP(&removeCols, "remove", "r", "", "Remove column(s) (comma-separated)")
	transformCmd.Flags().StringVarP(&renameCol, "rename", "n", "", "Rename column (format: old:new)")
	transformCmd.Flags().StringVarP(&updateCol, "update", "u", "", "Update column value (format: column=value)")
	if err := transformCmd.MarkFlagRequired("input"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
