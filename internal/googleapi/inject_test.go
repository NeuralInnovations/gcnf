package googleapi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gcnf/internal/config"
)

func TestInjectCommand_NonExistentFile(t *testing.T) {
	configs := &config.Configs{}
	err := InjectCommand("/nonexistent/path/template.env", false, "", configs)
	if err == nil {
		t.Fatal("expected error for non-existent template file, got nil")
	}
}

func TestInjectCommand_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "empty.env")

	if err := os.WriteFile(templatePath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	configs := &config.Configs{}
	err := InjectCommand(templatePath, false, "", configs)
	if err != nil {
		t.Fatalf("unexpected error for empty template: %v", err)
	}
}

func TestInjectCommand_PlainTextPassthrough(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "plain.env")

	content := "# This is a comment\nAPP_NAME=myapp\nAPP_PORT=8080\n"
	if err := os.WriteFile(templatePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	configs := &config.Configs{}
	// This should succeed — plain key=value lines with no gcnf:// URLs
	// are resolved with env var substitution only.
	err := InjectCommand(templatePath, false, "", configs)
	if err != nil {
		t.Fatalf("unexpected error for plain text template: %v", err)
	}
}

func TestInjectCommand_SkipComments(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "comments.env")

	content := "# comment line\nKEY=value\n// another comment\nKEY2=value2\n; semicolon comment\n"
	if err := os.WriteFile(templatePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	configs := &config.Configs{}
	// With skipComments=true, comment lines should be omitted from output.
	err := InjectCommand(templatePath, true, "", configs)
	if err != nil {
		t.Fatalf("unexpected error with skipComments=true: %v", err)
	}
}

func TestInjectCommand_EnvVarSubstitution(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "envvars.env")

	// Set a test environment variable.
	t.Setenv("GCNF_TEST_HOST", "testhost.example.com")

	content := "DB_HOST=$GCNF_TEST_HOST\nDB_PORT=5432\n"
	if err := os.WriteFile(templatePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	configs := &config.Configs{}
	err := InjectCommand(templatePath, false, "", configs)
	if err != nil {
		t.Fatalf("unexpected error for env var substitution: %v", err)
	}
	// The function prints to stdout, so we cannot easily capture the output
	// without redirecting os.Stdout. We verify it runs without error.
}

func TestInjectCommand_BracedEnvVarWithDefault(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "braced.env")

	// GCNF_TEST_UNDEFINED should not be set, so the default "fallback" is used.
	os.Unsetenv("GCNF_TEST_UNDEFINED")

	content := "VALUE=${GCNF_TEST_UNDEFINED:-fallback}\n"
	if err := os.WriteFile(templatePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	configs := &config.Configs{}
	err := InjectCommand(templatePath, false, "", configs)
	if err != nil {
		t.Fatalf("unexpected error for braced env var with default: %v", err)
	}
}

func TestInjectCommand_InternalVariableReference(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "internal.env")

	// The first line defines BASE_URL, the second references it via $BASE_URL.
	content := "BASE_URL=http://localhost:8080\nAPI_URL=$BASE_URL/api\n"
	if err := os.WriteFile(templatePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	configs := &config.Configs{}
	err := InjectCommand(templatePath, false, "", configs)
	if err != nil {
		t.Fatalf("unexpected error for internal variable reference: %v", err)
	}
}

func TestInjectCommand_MixedContentWithComments(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "mixed.env")

	content := `# Application Config
APP_NAME=myservice
APP_VERSION=1.0.0

# Database settings
DB_HOST=localhost
DB_PORT=5432
DB_NAME=mydb

# Derived value
DB_URL=$DB_HOST:$DB_PORT/$DB_NAME
`
	if err := os.WriteFile(templatePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	configs := &config.Configs{}
	err := InjectCommand(templatePath, false, "", configs)
	if err != nil {
		t.Fatalf("unexpected error for mixed content: %v", err)
	}

	// Also test with comments skipped.
	err = InjectCommand(templatePath, true, "", configs)
	if err != nil {
		t.Fatalf("unexpected error for mixed content with skipComments: %v", err)
	}
}

func TestInjectCommand_OutputToFile(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "template.env")
	outputPath := filepath.Join(tmpDir, "output.env")

	content := "APP_NAME=myapp\nAPP_PORT=8080\n"
	if err := os.WriteFile(templatePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	configs := &config.Configs{}
	err := InjectCommand(templatePath, false, outputPath, configs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	got := string(result)
	if !strings.Contains(got, "APP_NAME=myapp") || !strings.Contains(got, "APP_PORT=8080") {
		t.Errorf("output file missing expected content, got: %q", got)
	}
}

func TestInjectCommand_OutputToInvalidPath(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "template.env")

	if err := os.WriteFile(templatePath, []byte("KEY=value\n"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	configs := &config.Configs{}
	err := InjectCommand(templatePath, false, "/nonexistent/dir/output.env", configs)
	if err == nil {
		t.Fatal("expected error for invalid output path, got nil")
	}
}
