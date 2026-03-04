package googleapi

import (
	"strings"
	"testing"

	"gcnf/internal/config"
)

func TestReadGCNFURL_InvalidPrefix(t *testing.T) {
	configs := &config.Configs{}
	_, err := ReadGCNFURL("http://example.com/Sheet1/dev/Cat/Name", configs)
	if err == nil {
		t.Fatal("expected error for non-gcnf:// prefix, got nil")
	}
	if !strings.Contains(err.Error(), "must start with 'gcnf://'") {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestReadGCNFURL_MissingParts(t *testing.T) {
	configs := &config.Configs{}

	tests := []struct {
		name string
		url  string
	}{
		{"only sheet and env", "gcnf://Sheet1/dev"},
		{"only sheet, env, and category", "gcnf://Sheet1/dev/Category"},
		{"only sheet", "gcnf://Sheet1"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ReadGCNFURL(tc.url, configs)
			if err == nil {
				t.Fatalf("expected error for URL %q with missing parts, got nil", tc.url)
			}
			if !strings.Contains(err.Error(), "invalid gcnf URL format") {
				t.Errorf("unexpected error message for URL %q: %s", tc.url, err.Error())
			}
		})
	}
}

func TestReadGCNFURL_EmptyURL(t *testing.T) {
	configs := &config.Configs{}
	_, err := ReadGCNFURL("", configs)
	if err == nil {
		t.Fatal("expected error for empty URL, got nil")
	}
	if !strings.Contains(err.Error(), "must start with 'gcnf://'") {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestReadGCNFURL_RandomString(t *testing.T) {
	configs := &config.Configs{}
	_, err := ReadGCNFURL("some random string", configs)
	if err == nil {
		t.Fatal("expected error for random string, got nil")
	}
	if !strings.Contains(err.Error(), "must start with 'gcnf://'") {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

// Note: A URL like "gcnf://Sheet1/dev/Category/Name/Extra" passes the format
// check (SplitN with limit 4 folds extra slashes into the Name part), but then
// calls loadValue which invokes log.Fatalf/os.Exit when Google API is unavailable.
// We cannot test that case without mocking loadValue.

func TestReadGCNFURL_PrefixVariations(t *testing.T) {
	configs := &config.Configs{}

	// All these variations should fail the "gcnf://" prefix check.
	tests := []struct {
		name string
		url  string
	}{
		{"uppercase GCNF", "GCNF://Sheet1/dev/Cat/Name"},
		{"mixed case", "Gcnf://Sheet1/dev/Cat/Name"},
		{"missing colon", "gcnf//Sheet1/dev/Cat/Name"},
		{"single slash", "gcnf:/Sheet1/dev/Cat/Name"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ReadGCNFURL(tc.url, configs)
			if err == nil {
				t.Errorf("expected error for URL %q, got nil", tc.url)
			}
			if !strings.Contains(err.Error(), "must start with 'gcnf://'") {
				t.Errorf("unexpected error message for URL %q: %s", tc.url, err.Error())
			}
		})
	}
}
