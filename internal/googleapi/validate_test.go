package googleapi

import (
	"os"
	"path/filepath"
	"testing"

	"gcnf/internal/config"
)

func TestValidateCommand_ValidTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "valid.env")
	t.Setenv("GCNF_TEST_VAR", "hello")

	content := "APP_NAME=myapp\nAPP_HOST=$GCNF_TEST_VAR\n"
	os.WriteFile(templatePath, []byte(content), 0644)

	configs := &config.Configs{}
	issues := ValidateCommand(templatePath, configs)
	if len(issues) != 0 {
		t.Errorf("expected 0 issues for valid template, got %d: %v", len(issues), issues)
	}
}

func TestValidateCommand_MissingVariable(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "missing.env")
	os.Unsetenv("GCNF_UNDEFINED_XYZ")

	content := "KEY=$GCNF_UNDEFINED_XYZ\n"
	os.WriteFile(templatePath, []byte(content), 0644)

	configs := &config.Configs{}
	issues := ValidateCommand(templatePath, configs)
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d: %v", len(issues), issues)
	}
	if issues[0].Type != "missing_var" {
		t.Errorf("expected type 'missing_var', got %q", issues[0].Type)
	}
}

func TestValidateCommand_BracedMissingVariable(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "braced.env")
	os.Unsetenv("GCNF_UNDEF_ABC")

	content := "KEY=${GCNF_UNDEF_ABC}\n"
	os.WriteFile(templatePath, []byte(content), 0644)

	configs := &config.Configs{}
	issues := ValidateCommand(templatePath, configs)
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d: %v", len(issues), issues)
	}
	if issues[0].Type != "missing_var" {
		t.Errorf("expected type 'missing_var', got %q", issues[0].Type)
	}
}

func TestValidateCommand_BracedWithDefault_NoIssue(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "default.env")
	os.Unsetenv("GCNF_UNDEF_DEF")

	content := "KEY=${GCNF_UNDEF_DEF:-fallback}\n"
	os.WriteFile(templatePath, []byte(content), 0644)

	configs := &config.Configs{}
	issues := ValidateCommand(templatePath, configs)
	if len(issues) != 0 {
		t.Errorf("expected 0 issues for var with default, got %d: %v", len(issues), issues)
	}
}

func TestValidateCommand_MalformedGcnfURL(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "malformed.env")

	content := "KEY=gcnf://Sheet/Env\n"
	os.WriteFile(templatePath, []byte(content), 0644)

	configs := &config.Configs{}
	issues := ValidateCommand(templatePath, configs)
	found := false
	for _, issue := range issues {
		if issue.Type == "malformed_url" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected malformed_url issue, got: %v", issues)
	}
}

func TestValidateCommand_ValidGcnfURL(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "valid_url.env")

	content := "KEY=gcnf://Sheet/Env/Category/Name\n"
	os.WriteFile(templatePath, []byte(content), 0644)

	configs := &config.Configs{}
	issues := ValidateCommand(templatePath, configs)
	for _, issue := range issues {
		if issue.Type == "malformed_url" {
			t.Errorf("unexpected malformed_url issue for valid URL: %v", issue)
		}
	}
}

func TestValidateCommand_EmptyValue(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "empty.env")

	content := "KEY=\n"
	os.WriteFile(templatePath, []byte(content), 0644)

	configs := &config.Configs{}
	issues := ValidateCommand(templatePath, configs)
	found := false
	for _, issue := range issues {
		if issue.Type == "empty_value" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected empty_value issue, got: %v", issues)
	}
}

func TestValidateCommand_NonExistentFile(t *testing.T) {
	configs := &config.Configs{}
	issues := ValidateCommand("/nonexistent/template.env", configs)
	if len(issues) != 1 || issues[0].Type != "error" {
		t.Errorf("expected 1 error issue for non-existent file, got: %v", issues)
	}
}

func TestValidateCommand_DefinedVariableNotFlagged(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "defined.env")

	content := "BASE=http://localhost\nURL=$BASE/api\n"
	os.WriteFile(templatePath, []byte(content), 0644)

	configs := &config.Configs{}
	issues := ValidateCommand(templatePath, configs)
	for _, issue := range issues {
		if issue.Type == "missing_var" {
			t.Errorf("should not flag $BASE as missing since it was defined earlier: %v", issue)
		}
	}
}

func TestValidateCommand_CommentsSkipped(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "comments.env")

	content := "# This is a comment\n// Another comment\nKEY=value\n"
	os.WriteFile(templatePath, []byte(content), 0644)

	configs := &config.Configs{}
	issues := ValidateCommand(templatePath, configs)
	if len(issues) != 0 {
		t.Errorf("expected 0 issues for template with comments, got %d: %v", len(issues), issues)
	}
}
