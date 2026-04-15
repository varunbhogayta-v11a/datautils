package cmd

import (
	"fmt"
	"os"

	"github.com/improwised/datautil/pkg/data"
	"github.com/spf13/cobra"
)

var (
	toFormat    string
	prettyPrint bool
	showStats   bool
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export dataset to different format",
	Long: `Export data to different formats with optional transformations.

Examples:
  datautil export --input data.csv --to json --output result.json --token "YOUR_JWT"
  datautil export --input data.json --to xml --output result.xml --token "YOUR_JWT"`,
	Run: func(cmd *cobra.Command, args []string) {
		requireAuth(cmd)
		ensureDB()

		ds, err := data.ReadFile(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}

		if showStats {
			fmt.Printf("Dataset Statistics:\n")
			fmt.Printf("  Rows: %d\n", ds.RowCount())
			fmt.Printf("  Columns: %d\n", ds.ColCount())
			fmt.Printf("  Headers: %v\n", ds.Headers)
			fmt.Println()
		}

		if outputFile != "" {
			if err := data.WriteFile(ds, outputFile, toFormat); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Exported to %s (%d rows)\n", outputFile, ds.RowCount())
		} else {
			fmt.Printf("Output format: %s\n", toFormat)
			fmt.Println("(Use --output to save to file)")
		}
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input file (required)")
	exportCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file")
	exportCmd.Flags().StringVarP(&toFormat, "to", "t", "", "Output format: csv, json, xml (required)")
	exportCmd.Flags().BoolVarP(&prettyPrint, "pretty", "p", false, "Pretty print JSON/XML output")
	exportCmd.Flags().BoolVarP(&showStats, "stats", "s", false, "Show summary statistics")
	if err := exportCmd.MarkFlagRequired("input"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := exportCmd.MarkFlagRequired("to"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
