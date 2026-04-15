package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/improwised/datautil/pkg/data"
	"github.com/improwised/datautil/pkg/operations"
	"github.com/spf13/cobra"
)

var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Launch interactive mode",
	Long:  `Launch an interactive menu for dataset processing`,
	Run: func(cmd *cobra.Command, args []string) {
		runInteractive()
	},
}

func runInteractive() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== DataUtil Interactive Mode ===")
	fmt.Println()

	fmt.Print("Enter input file: ")
	inputFile, _ := reader.ReadString('\n')
	inputFile = strings.TrimSpace(inputFile)

	if inputFile == "" {
		fmt.Println("Error: No input file specified")
		return
	}

	ds, err := data.ReadFile(inputFile)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	fmt.Printf("Loaded %d rows, %d columns\n", ds.RowCount(), ds.ColCount())
	fmt.Println("Columns:", strings.Join(ds.Headers, ", "))
	fmt.Println()

	fmt.Println("Available operations:")
	fmt.Println("  1. Filter rows")
	fmt.Println("  2. Select columns")
	fmt.Println("  3. Add column")
	fmt.Println("  4. Remove column")
	fmt.Println("  5. Validate")
	fmt.Println("  6. Export")
	fmt.Println("  7. Show data")
	fmt.Println("  8. Quit")
	fmt.Println()

	fmt.Print("Choose operation (1-8): ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		fmt.Print("Enter filter condition (e.g., age > 25): ")
		cond, _ := reader.ReadString('\n')
		ds, err = operations.FilterRows(ds, strings.TrimSpace(cond), false)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("Filtered to %d rows\n", ds.RowCount())

	case "2":
		fmt.Print("Enter columns (comma-separated): ")
		cols, _ := reader.ReadString('\n')
		ds = operations.SelectColumns(ds, strings.Split(strings.TrimSpace(cols), ","))
		fmt.Printf("Selected %d columns\n", ds.ColCount())

	case "3":
		fmt.Print("Enter new column name: ")
		name, _ := reader.ReadString('\n')
		fmt.Print("Enter expression: ")
		expr, _ := reader.ReadString('\n')
		ds = operations.AddColumn(ds, strings.TrimSpace(name), strings.TrimSpace(expr))
		fmt.Printf("Added column: %s\n", strings.TrimSpace(name))

	case "4":
		fmt.Print("Enter columns to remove (comma-separated): ")
		cols, _ := reader.ReadString('\n')
		ds = operations.RemoveColumns(ds, strings.Split(strings.TrimSpace(cols), ","))
		fmt.Printf("Removed columns, %d remaining\n", ds.ColCount())

	case "5":
		fmt.Print("Enter required columns (comma-separated): ")
		req, _ := reader.ReadString('\n')
		result := operations.ValidateDataset(ds, strings.Split(strings.TrimSpace(req), ","), nil)
		if result.Valid {
			fmt.Println("✓ Validation passed")
		} else {
			fmt.Println("✗ Validation failed:")
			for _, e := range result.Errors {
				fmt.Printf("  - %s\n", e)
			}
		}

	case "6":
		fmt.Print("Enter output file: ")
		out, _ := reader.ReadString('\n')
		fmt.Print("Enter format (csv, json, xml): ")
		fmtf, _ := reader.ReadString('\n')
		data.WriteFile(ds, strings.TrimSpace(out), strings.TrimSpace(fmtf))
		fmt.Printf("Saved to %s\n", strings.TrimSpace(out))

	case "7":
		data.PrintDataset(ds)

	case "8":
		fmt.Println("Goodbye!")
		return

	default:
		fmt.Println("Invalid choice")
	}
}

func init() {
	rootCmd.AddCommand(interactiveCmd)
}
