package operations

import (
	"testing"

	"github.com/varunbhogayta-v11a/datautils/pkg/data"
)

func TestAddColumn(t *testing.T) {
	ds := &data.Dataset{
		Headers: []string{"name", "age"},
		Rows: [][]string{
			{"John", "25"},
			{"Jane", "30"},
		},
	}

	result := AddColumn(ds, "city", "NYC")

	if result.ColCount() != 3 {
		t.Errorf("ColCount() = %d, want 3", result.ColCount())
	}

	if result.Headers[2] != "city" {
		t.Errorf("Headers[2] = %q, want %q", result.Headers[2], "city")
	}

	if len(result.Rows) != 2 {
		t.Errorf("Row count = %d, want 2", len(result.Rows))
	}
}

func TestAddColumnWithExpression(t *testing.T) {
	ds := &data.Dataset{
		Headers: []string{"first_name", "last_name"},
		Rows: [][]string{
			{"John", "Doe"},
		},
	}

	result := AddColumn(ds, "full_name", "first_name+last_name")

	if result.ColCount() != 3 {
		t.Errorf("ColCount() = %d, want 3", result.ColCount())
	}
}

func TestFilterRows_EdgeCases(t *testing.T) {
	ds := &data.Dataset{
		Headers: []string{"name", "age"},
		Rows: [][]string{
			{"John", "25"},
			{"Jane", "notanumber"},
			{"Bob", "35"},
		},
	}

	t.Run("invalid column", func(t *testing.T) {
		_, err := FilterRows(ds, "invalidcol > 25", false)
		if err == nil {
			t.Error("expected error for invalid column")
		}
	})

	t.Run("invalid condition format", func(t *testing.T) {
		_, err := FilterRows(ds, "age", false)
		if err == nil {
			t.Error("expected error for invalid condition")
		}
	})

	t.Run("empty dataset", func(t *testing.T) {
		emptyDs := &data.Dataset{
			Headers: []string{"name"},
			Rows:    [][]string{},
		}
		result, err := FilterRows(emptyDs, "name == John", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.RowCount() != 0 {
			t.Errorf("RowCount() = %d, want 0", result.RowCount())
		}
	})

	t.Run("row shorter than header", func(t *testing.T) {
		shortRowDs := &data.Dataset{
			Headers: []string{"name", "age", "city"},
			Rows: [][]string{
				{"John"},
			},
		}
		result, err := FilterRows(shortRowDs, "age > 25", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.RowCount() != 0 {
			t.Errorf("RowCount() = %d, want 0", result.RowCount())
		}
	})
}

func TestSelectColumns_NotFound(t *testing.T) {
	ds := &data.Dataset{
		Headers: []string{"name", "age"},
		Rows: [][]string{
			{"John", "25"},
		},
	}

	result := SelectColumns(ds, []string{"nonexistent"})

	// When no columns match, returns original dataset unchanged
	if result.ColCount() != 2 {
		t.Errorf("ColCount() = %d, want 2 (original)", result.ColCount())
	}
}

func TestRemoveColumns_NotFound(t *testing.T) {
	ds := &data.Dataset{
		Headers: []string{"name", "age"},
		Rows: [][]string{
			{"John", "25"},
		},
	}

	result := RemoveColumns(ds, []string{"nonexistent"})

	if result.ColCount() != 2 {
		t.Errorf("ColCount() = %d, want 2", result.ColCount())
	}
}

func TestRenameColumn_NotFound(t *testing.T) {
	ds := &data.Dataset{
		Headers: []string{"name", "age"},
		Rows: [][]string{
			{"John", "25"},
		},
	}

	result := RenameColumn(ds, "nonexistent", "new_name")

	if result.Headers[0] != "name" {
		t.Errorf("Headers[0] = %q, want %q", result.Headers[0], "name")
	}
}

func TestValidateDataset_TypeErrors(t *testing.T) {
	ds := &data.Dataset{
		Headers: []string{"name", "age"},
		Rows: [][]string{
			{"John", "25"},
			{"Jane", "notanumber"},
		},
	}

	result := ValidateDataset(ds, []string{"name"}, map[string]string{"age": "int"})

	if result.Valid {
		t.Error("expected validation to fail for invalid type")
	}

	if len(result.Errors) == 0 {
		t.Error("expected errors to be reported")
	}
}

func TestValidateDataset_EmptyDataset(t *testing.T) {
	ds := &data.Dataset{
		Headers: []string{},
		Rows:    [][]string{},
	}

	result := ValidateDataset(ds, []string{}, nil)

	if !result.Valid {
		t.Error("expected validation to pass for empty dataset with no requirements")
	}
}

func TestParseCondition(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"age > 25", []string{"age", ">", "25"}},
		{"name == John", []string{"name", "==", "John"}},
		{"city != NYC", []string{"city", "!=", "NYC"}},
		{"score >= 100", []string{"score", ">=", "100"}},
		{"count <= 10", []string{"count", "<=", "10"}},
		{"text contains hello", []string{"text", "contains", "hello"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseCondition(tt.input)
			if len(got) != len(tt.want) || got[0] != tt.want[0] || got[1] != tt.want[1] || got[2] != tt.want[2] {
				t.Errorf("parseCondition(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestEvaluateCondition(t *testing.T) {
	tests := []struct {
		rowVal    string
		operator  string
		targetVal string
		want      bool
	}{
		{"25", ">", "20", true},
		{"15", ">", "20", false},
		{"20", "==", "20", true},
		{"20", "!=", "25", true},
		{"hello", "contains", "ell", true},
		{"hello", "contains", "xyz", false},
		{"100", ">=", "100", true},
		{"50", "<=", "100", true},
		{"abc", ">", "def", false},
	}

	for _, tt := range tests {
		t.Run(tt.operator, func(t *testing.T) {
			got := evaluateCondition(tt.rowVal, tt.operator, tt.targetVal)
			if got != tt.want {
				t.Errorf("evaluateCondition(%q, %q, %q) = %v, want %v", tt.rowVal, tt.operator, tt.targetVal, got, tt.want)
			}
		})
	}
}

func TestCompareNumeric(t *testing.T) {
	tests := []struct {
		a    string
		b    string
		want int
	}{
		{"25", "20", 1},
		{"20", "25", -1},
		{"20", "20", 0},
		{"abc", "def", -1},
		{"10.5", "10.5", 0},
		{"10.5", "20.0", -1},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			got := compareNumeric(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("compareNumeric(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
