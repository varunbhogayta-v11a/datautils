package operations

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/improwised/datautil/pkg/data"
)

func FilterRows(ds *data.Dataset, condition string, invert bool) (*data.Dataset, error) {
	parts := parseCondition(condition)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid condition format. Use: column operator value (e.g., 'age > 25')")
	}

	colName := strings.TrimSpace(parts[0])
	operator := strings.TrimSpace(parts[1])
	value := strings.TrimSpace(parts[2])

	colIndex := -1
	for i, h := range ds.Headers {
		if strings.EqualFold(h, colName) {
			colIndex = i
			break
		}
	}

	if colIndex == -1 {
		return nil, fmt.Errorf("column '%s' not found", colName)
	}

	filteredRows := make([][]string, 0)
	for _, row := range ds.Rows {
		if colIndex >= len(row) {
			continue
		}
		rowVal := row[colIndex]
		matches := evaluateCondition(rowVal, operator, value)

		if invert {
			matches = !matches
		}

		if matches {
			filteredRows = append(filteredRows, row)
		}

	}

	return &data.Dataset{Headers: ds.Headers, Rows: filteredRows}, nil
}

func parseCondition(cond string) []string {
	re := regexp.MustCompile(`(\S+)\s*(==|!=|>=|<=|>|<|=|contains)\s*(\S+)`)
	matches := re.FindStringSubmatch(cond)
	if matches == nil {
		parts := strings.SplitN(cond, " ", 3)
		return parts
	}
	return matches[1:]
}

func evaluateCondition(rowVal, operator, targetVal string) bool {
	switch operator {
	case "==", "=":
		return rowVal == targetVal
	case "!=":
		return rowVal != targetVal
	case ">":
		return compareNumeric(rowVal, targetVal) > 0
	case "<":
		return compareNumeric(rowVal, targetVal) < 0
	case ">=":
		return compareNumeric(rowVal, targetVal) >= 0
	case "<=":
		return compareNumeric(rowVal, targetVal) <= 0
	case "contains":
		return strings.Contains(rowVal, targetVal)
	default:
		return rowVal == targetVal
	}
}

func compareNumeric(a, b string) int {
	aFloat, aErr := strconv.ParseFloat(a, 64)
	bFloat, bErr := strconv.ParseFloat(b, 64)

	if aErr != nil || bErr != nil {
		return strings.Compare(a, b)
	}

	if aFloat < bFloat {
		return -1
	} else if aFloat > bFloat {
		return 1
	}
	return 0
}

func SelectColumns(ds *data.Dataset, columns []string) *data.Dataset {
	colIndices := make([]int, 0)
	newHeaders := make([]string, 0)

	for _, col := range columns {
		col = strings.TrimSpace(col)
		for i, h := range ds.Headers {
			if strings.EqualFold(h, col) {
				colIndices = append(colIndices, i)
				newHeaders = append(newHeaders, h)
				break
			}
		}
	}

	if len(colIndices) == 0 {
		return ds
	}

	newRows := make([][]string, len(ds.Rows))
	for i, row := range ds.Rows {
		newRows[i] = make([]string, len(colIndices))
		for j, idx := range colIndices {
			if idx < len(row) {
				newRows[i][j] = row[idx]
			}
		}
	}

	return &data.Dataset{Headers: newHeaders, Rows: newRows}
}

func AddColumn(ds *data.Dataset, name string, expression string) *data.Dataset {
	newHeaders := append(ds.Headers, name)
	newRows := make([][]string, len(ds.Rows))

	for i, row := range ds.Rows {
		newRow := make([]string, len(row)+1)
		copy(newRow, row)
		newRow[len(row)] = evaluateExpression(row, ds.Headers, expression)
		newRows[i] = newRow
	}

	return &data.Dataset{Headers: newHeaders, Rows: newRows}
}

func evaluateExpression(row []string, headers []string, expr string) string {
	if strings.Contains(expr, "+") {
		parts := strings.Split(expr, "+")
		result := ""
		for _, p := range parts {
			p = strings.TrimSpace(p)
			for j, h := range headers {
				if strings.EqualFold(h, p) && j < len(row) {
					result += row[j]
					break
				}
			}
			if result == "" {
				result += p
			}
		}
		return result
	}
	return expr
}

func RemoveColumns(ds *data.Dataset, columns []string) *data.Dataset {
	removeIdx := make(map[int]bool)
	for _, col := range columns {
		col = strings.TrimSpace(col)
		for i, h := range ds.Headers {
			if strings.EqualFold(h, col) {
				removeIdx[i] = true
				break
			}
		}
	}

	newHeaders := make([]string, 0)
	for i, h := range ds.Headers {
		if !removeIdx[i] {
			newHeaders = append(newHeaders, h)
		}
	}

	newRows := make([][]string, len(ds.Rows))
	for i, row := range ds.Rows {
		newRow := make([]string, 0)
		for j, val := range row {
			if !removeIdx[j] {
				newRow = append(newRow, val)
			}
		}
		newRows[i] = newRow
	}

	return &data.Dataset{Headers: newHeaders, Rows: newRows}
}

func RenameColumn(ds *data.Dataset, oldName, newName string) *data.Dataset {
	newHeaders := make([]string, len(ds.Headers))
	for i, h := range ds.Headers {
		if strings.EqualFold(h, oldName) {
			newHeaders[i] = newName
		} else {
			newHeaders[i] = h
		}
	}

	return &data.Dataset{Headers: newHeaders, Rows: ds.Rows}
}

type ValidationResult struct {
	Valid  bool
	Errors []string
}

func ValidateDataset(ds *data.Dataset, requiredCols []string, types map[string]string) *ValidationResult {
	result := &ValidationResult{Valid: true, Errors: []string{}}

	for _, col := range requiredCols {
		found := false
		for _, h := range ds.Headers {
			if strings.EqualFold(h, col) {
				found = true
				break
			}
		}
		if !found {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("Required column '%s' not found", col))
		}
	}

	for col, expectedType := range types {
		colIdx := -1
		for i, h := range ds.Headers {
			if strings.EqualFold(h, col) {
				colIdx = i
				break
			}
		}
		if colIdx == -1 {
			continue
		}

		for rowIdx, row := range ds.Rows {
			if colIdx >= len(row) {
				continue
			}
			val := row[colIdx]
			if !validateType(val, expectedType) {
				result.Valid = false
				result.Errors = append(result.Errors, fmt.Sprintf("Row %d: '%s' is not of type %s", rowIdx+1, col, expectedType))
			}
		}
	}

	return result
}

func validateType(value, expectedType string) bool {
	if value == "" {
		return true
	}

	switch strings.ToLower(expectedType) {
	case "string":
		return true
	case "int", "integer":
		_, err := strconv.Atoi(value)
		return err == nil
	case "float", "number":
		_, err := strconv.ParseFloat(value, 64)
		return err == nil
	case "bool", "boolean":
		_, err := strconv.ParseBool(value)
		return err == nil
	default:
		return true
	}
}

var _ = reflect.TypeOf(nil)
