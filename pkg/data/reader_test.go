package data

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"data.csv", "csv"},
		{"data.json", "json"},
		{"data.xml", "xml"},
		{"data.xlsx", "excel"},
		{"data.xls", "excel"},
		{"data.txt", "unknown"},
		{"data", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			if got := DetectFormat(tt.filename); got != tt.want {
				t.Errorf("DetectFormat(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestReadCSV(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.csv")
	content := "name,age,city\nJohn,25,NYC\nJane,30,LA\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	ds, err := ReadCSV(testFile)
	if err != nil {
		t.Fatalf("ReadCSV() error = %v", err)
	}

	if ds.ColCount() != 3 {
		t.Errorf("ColCount() = %d, want 3", ds.ColCount())
	}

	if ds.RowCount() != 2 {
		t.Errorf("RowCount() = %d, want 2", ds.RowCount())
	}

	if ds.Headers[0] != "name" {
		t.Errorf("Headers[0] = %q, want %q", ds.Headers[0], "name")
	}
}

func TestReadJSON(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")
	content := `[{"name":"John","age":25},{"name":"Jane","age":30}]`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	ds, err := ReadJSON(testFile)
	if err != nil {
		t.Fatalf("ReadJSON() error = %v", err)
	}

	if ds.RowCount() != 2 {
		t.Errorf("RowCount() = %d, want 2", ds.RowCount())
	}
}

func TestReadXML(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.xml")
	content := `<?xml version="1.0"?><data><record><name>John</name><age>25</age></record></data>`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	ds, err := ReadXML(testFile)
	if err != nil {
		t.Fatalf("ReadXML() error = %v", err)
	}

	if ds.RowCount() != 1 {
		t.Errorf("RowCount() = %d, want 1", ds.RowCount())
	}
}

func TestReadFile(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("unsupported format", func(t *testing.T) {
		_, err := ReadFile(filepath.Join(tmpDir, "test.txt"))
		if err == nil {
			t.Error("expected error for unsupported format")
		}
	})

	t.Run("csv file", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "test.csv")
		os.WriteFile(testFile, []byte("name,age\nJohn,25\n"), 0644)
		ds, err := ReadFile(testFile)
		if err != nil {
			t.Fatalf("ReadFile() error = %v", err)
		}
		if ds.ColCount() != 2 {
			t.Errorf("ColCount() = %d, want 2", ds.ColCount())
		}
	})

	t.Run("json file", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "test.json")
		os.WriteFile(testFile, []byte(`[{"name":"John"}]`), 0644)
		ds, err := ReadFile(testFile)
		if err != nil {
			t.Fatalf("ReadFile() error = %v", err)
		}
		if ds.RowCount() != 1 {
			t.Errorf("RowCount() = %d, want 1", ds.RowCount())
		}
	})

	t.Run("xml file", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "test.xml")
		os.WriteFile(testFile, []byte(`<?xml?><data><record><name>John</name></record></data>`), 0644)
		ds, err := ReadFile(testFile)
		if err != nil {
			t.Fatalf("ReadFile() error = %v", err)
		}
		if ds.RowCount() != 1 {
			t.Errorf("RowCount() = %d, want 1", ds.RowCount())
		}
	})
}

func TestWriteCSV(t *testing.T) {
	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "output.csv")

	ds := &Dataset{
		Headers: []string{"name", "age"},
		Rows:    [][]string{{"John", "25"}, {"Jane", "30"}},
	}

	if err := WriteCSV(ds, outFile); err != nil {
		t.Fatalf("WriteCSV() error = %v", err)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	expected := "name,age\nJohn,25\nJane,30\n"
	if string(data) != expected {
		t.Errorf("got %q, want %q", string(data), expected)
	}
}

func TestWriteJSON(t *testing.T) {
	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "output.json")

	ds := &Dataset{
		Headers: []string{"name", "age"},
		Rows:    [][]string{{"John", "25"}},
	}

	if err := WriteJSON(ds, outFile); err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestWriteXML(t *testing.T) {
	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "output.xml")

	ds := &Dataset{
		Headers: []string{"name", "age"},
		Rows:    [][]string{{"John", "25"}},
	}

	if err := WriteXML(ds, outFile); err != nil {
		t.Fatalf("WriteXML() error = %v", err)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestDatasetRowCount(t *testing.T) {
	ds := &Dataset{
		Headers: []string{"a", "b"},
		Rows:    [][]string{{"1", "2"}, {"3", "4"}},
	}

	if ds.RowCount() != 2 {
		t.Errorf("RowCount() = %d, want 2", ds.RowCount())
	}
}

func TestDatasetColCount(t *testing.T) {
	ds := &Dataset{
		Headers: []string{"a", "b", "c"},
		Rows:    [][]string{{"1", "2", "3"}},
	}

	if ds.ColCount() != 3 {
		t.Errorf("ColCount() = %d, want 3", ds.ColCount())
	}
}

func TestReadFile_NotFound(t *testing.T) {
	_, err := ReadFile("/nonexistent/path/file.csv")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestWriteFile_Unsupported(t *testing.T) {
	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "output.txt")

	ds := &Dataset{
		Headers: []string{"name"},
		Rows:    [][]string{{"John"}},
	}

	err := WriteFile(ds, outFile, "txt")
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestPrintDataset(t *testing.T) {
	ds := &Dataset{
		Headers: []string{"name", "age"},
		Rows:    [][]string{{"John", "25"}},
	}

	PrintDataset(ds)
}

func TestWriteFile_WithFormat(t *testing.T) {
	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "output")

	ds := &Dataset{
		Headers: []string{"name", "age"},
		Rows:    [][]string{{"John", "25"}},
	}

	err := WriteFile(ds, outFile, "csv")
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err = os.ReadFile(outFile)
	if err != nil {
		t.Errorf("expected output file: %v", err)
	}
}

func TestWriteFile_PreserveFormat(t *testing.T) {
	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "output.csv")

	ds := &Dataset{
		Headers: []string{"name"},
		Rows:    [][]string{{"John"}},
	}

	err := WriteFile(ds, outFile, "")
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}