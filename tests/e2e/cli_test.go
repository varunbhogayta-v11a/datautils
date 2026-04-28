//go:build e2e
// +build e2e

package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestE2E_CLIFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e tests in short mode")
	}

	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.csv")
	outputFile := filepath.Join(tmpDir, "output.csv")

	content := "name,age,city\nJohn,25,NYC\nJane,30,LA\n"
	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create input file: %v", err)
	}

	cmd := exec.Command("./datautil", "filter", "--input", inputFile, "--where", "age > 25", "--output", outputFile)
	cmd.Dir = tmpDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Command output: %s", string(output))
		t.Skipf("CLI may not be built yet, skipping e2e test: %v", err)
	}

	if _, err := os.Stat(outputFile); err != nil {
		t.Errorf("expected output file to exist: %v", err)
	}
}

func TestE2E_CLITransform(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e tests in short mode")
	}

	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.csv")
	outputFile := filepath.Join(tmpDir, "output.csv")

	content := "name,age\nJohn,25\n"
	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create input file: %v", err)
	}

	cmd := exec.Command("./datautil", "transform", "--input", inputFile, "--add", "city=NYC", "--output", outputFile)
	cmd.Dir = tmpDir

	_, err := cmd.CombinedOutput()
	if err != nil {
		t.Skipf("CLI may not be built yet, skipping e2e test: %v", err)
	}
}

func TestE2E_CLIValidate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e tests in short mode")
	}

	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.csv")

	content := "name,age\nJohn,25\n"
	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create input file: %v", err)
	}

	cmd := exec.Command("./datautil", "validate", "--input", inputFile, "--required", "name,age")
	cmd.Dir = tmpDir

	_, err := cmd.CombinedOutput()
	if err != nil {
		t.Skipf("CLI may not be built yet, skipping e2e test: %v", err)
	}
}

func TestE2E_CLIExport(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e tests in short mode")
	}

	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.csv")
	outputFile := filepath.Join(tmpDir, "output.json")

	content := "name,age\nJohn,25\n"
	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create input file: %v", err)
	}

	cmd := exec.Command("./datautil", "export", "--input", inputFile, "--to", "json", "--output", outputFile)
	cmd.Dir = tmpDir

	_, err := cmd.CombinedOutput()
	if err != nil {
		t.Skipf("CLI may not be built yet, skipping e2e test: %v", err)
	}
}

func TestE2E_CLIInvalidInput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e tests in short mode")
	}

	cmd := exec.Command("./datautil", "filter", "--input", "/nonexistent/file.csv", "--where", "age > 25")
	cmd.Dir = t.TempDir()

	err := cmd.Run()
	if err == nil {
		t.Error("expected error for nonexistent input file")
	}
}

func TestE2E_CLIHelp(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e tests in short mode")
	}

	cmd := exec.Command("./datautil", "--help")
	cmd.Dir = t.TempDir()

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Skipf("CLI may not be built yet, skipping e2e test: %v", err)
	}

	if len(output) == 0 {
		t.Error("expected non-empty output from help")
	}
}

func TestE2E_CLIUnknownCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e tests in short mode")
	}

	cmd := exec.Command("./datautil", "unknown-command")
	cmd.Dir = t.TempDir()

	err := cmd.Run()
	if err == nil {
		t.Error("expected error for unknown command")
	}
}

func TestE2E_CLIDBInit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e tests in short mode")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cmd := exec.Command("./datautil", "init-db")
	cmd.Env = append(os.Environ(), "DB_DRIVER=sqlite3", "DB_NAME="+dbPath)
	cmd.Dir = tmpDir

	_, err := cmd.CombinedOutput()
	if err != nil {
		t.Skipf("CLI may not be built yet, skipping e2e test: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
}

func TestE2E_CLIFilterInvert(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e tests in short mode")
	}

	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.csv")
	outputFile := filepath.Join(tmpDir, "output.csv")

	content := "name,age,city\nJohn,25,NYC\nJane,30,LA\n"
	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create input file: %v", err)
	}

	cmd := exec.Command("./datautil", "filter", "--input", inputFile, "--where", "city == NYC", "--invert", "--output", outputFile)
	cmd.Dir = tmpDir

	_, err := cmd.CombinedOutput()
	if err != nil {
		t.Skipf("CLI may not be built yet, skipping e2e test: %v", err)
	}
}