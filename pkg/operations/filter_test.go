package operations

import (
	"testing"

	"github.com/varunbhogayta-v11a/datautils/pkg/data"
)

func TestFilterRows(t *testing.T) {
	ds := &data.Dataset{
		Headers: []string{"name", "age", "city"},
		Rows: [][]string{
			{"John", "25", "NYC"},
			{"Jane", "30", "LA"},
			{"Bob", "35", "NYC"},
		},
	}

	tests := []struct {
		name     string
		cond     string
		invert   bool
		wantRows int
	}{
		{"age > 25", "age > 25", false, 2},
		{"age < 30", "age < 30", false, 1},
		{"city == NYC", "city == NYC", false, 2},
		{"invert city == NYC", "city == NYC", true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FilterRows(ds, tt.cond, tt.invert)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.RowCount() != tt.wantRows {
				t.Errorf("got %d rows, want %d", result.RowCount(), tt.wantRows)
			}
		})
	}
}

func TestSelectColumns(t *testing.T) {
	ds := &data.Dataset{
		Headers: []string{"name", "age", "city"},
		Rows: [][]string{
			{"John", "25", "NYC"},
			{"Jane", "30", "LA"},
		},
	}

	result := SelectColumns(ds, []string{"name", "age"})

	if result.ColCount() != 2 {
		t.Errorf("got %d cols, want 2", result.ColCount())
	}

	if result.Headers[0] != "name" || result.Headers[1] != "age" {
		t.Errorf("unexpected headers: %v", result.Headers)
	}
}

func TestRemoveColumns(t *testing.T) {
	ds := &data.Dataset{
		Headers: []string{"name", "age", "city"},
		Rows: [][]string{
			{"John", "25", "NYC"},
		},
	}

	result := RemoveColumns(ds, []string{"age"})

	if result.ColCount() != 2 {
		t.Errorf("got %d cols, want 2", result.ColCount())
	}

	if result.Headers[0] != "name" || result.Headers[1] != "city" {
		t.Errorf("unexpected headers: %v", result.Headers)
	}
}

func TestRenameColumn(t *testing.T) {
	ds := &data.Dataset{
		Headers: []string{"name", "age"},
		Rows: [][]string{
			{"John", "25"},
		},
	}

	result := RenameColumn(ds, "name", "full_name")

	if result.Headers[0] != "full_name" {
		t.Errorf("got %s, want full_name", result.Headers[0])
	}
}

func TestValidateDataset(t *testing.T) {
	ds := &data.Dataset{
		Headers: []string{"name", "age"},
		Rows: [][]string{
			{"John", "25"},
			{"Jane", "30"},
		},
	}

	result := ValidateDataset(ds, []string{"name", "age"}, map[string]string{"age": "int"})

	if !result.Valid {
		t.Errorf("expected validation to pass")
	}

	result = ValidateDataset(ds, []string{"name", "email"}, nil)
	if result.Valid {
		t.Errorf("expected validation to fail for missing required column")
	}
}
