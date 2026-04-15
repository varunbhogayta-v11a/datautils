package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/improwised/datautil/pkg/auth"
	"github.com/improwised/datautil/pkg/data"
	"github.com/improwised/datautil/pkg/operations"
	"github.com/spf13/cobra"
)

var (
	inputFile    string
	outputFile   string
	format       string
	whereCond    string
	selectCols   string
	invertFilter bool
)

var filterCmd = &cobra.Command{
	Use:   "filter",
	Short: "Filter rows and columns from dataset",
	Long: `Filter data from input file based on conditions.

Examples:
  datautil filter --input data.csv --where "age > 25" --token "YOUR_JWT"
  datautil filter --input data.json --select name,email --output result.json`,
	Run: func(cmd *cobra.Command, args []string) {
		requireAuth(cmd)
		ensureDB()

		ds, err := data.ReadFile(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}

		if whereCond != "" {
			ds, err = operations.FilterRows(ds, whereCond, invertFilter)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error filtering rows: %v\n", err)
				os.Exit(1)
			}
		}

		if selectCols != "" {
			ds = operations.SelectColumns(ds, strings.Split(selectCols, ","))
		}

		if ds.RowCount() == 0 {
			fmt.Println("No matching records found")
			os.Exit(0)
		}

		if outputFile != "" {
			if err := data.WriteFile(ds, outputFile, format); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Filtered data saved to %s (%d rows)\n", outputFile, ds.RowCount())
			auth.LogOperation(currentUser.UserID, "filter", inputFile, outputFile, whereCond)
		} else {
			data.PrintDataset(ds)
		}
	},
}

func init() {
	rootCmd.AddCommand(filterCmd)
	filterCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input file (required)")
	filterCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file")
	filterCmd.Flags().StringVarP(&format, "format", "f", "", "Output format (csv, json, xml)")
	filterCmd.Flags().StringVarP(&whereCond, "where", "w", "", "Filter condition (e.g., age > 25)")
	filterCmd.Flags().StringVarP(&selectCols, "select", "s", "", "Select columns (comma-separated)")
	filterCmd.Flags().BoolVarP(&invertFilter, "invert", "", false, "Invert filter (exclude matching rows)")
	if err := filterCmd.MarkFlagRequired("input"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
