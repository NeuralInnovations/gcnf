package googleapi

import (
	"testing"
)

// ---------------------------------------------------------------------------
// Tests for extractEnvNames
// ---------------------------------------------------------------------------

func TestExtractEnvNames_Normal(t *testing.T) {
	header := []interface{}{"Category", "Name", "dev", "staging", "production"}
	envs := extractEnvNames(header)
	expected := []string{"dev", "staging", "production"}
	if len(envs) != len(expected) {
		t.Fatalf("expected %d envs, got %d", len(expected), len(envs))
	}
	for i, e := range expected {
		if envs[i] != e {
			t.Errorf("envs[%d] = %q, want %q", i, envs[i], e)
		}
	}
}

func TestExtractEnvNames_OnlyTwoColumns(t *testing.T) {
	header := []interface{}{"Category", "Name"}
	envs := extractEnvNames(header)
	if len(envs) != 0 {
		t.Errorf("expected 0 envs for 2-column header, got %d", len(envs))
	}
}

func TestExtractEnvNames_NonStringValues(t *testing.T) {
	header := []interface{}{"Category", "Name", 123, true}
	envs := extractEnvNames(header)
	if len(envs) != 2 {
		t.Fatalf("expected 2 envs, got %d", len(envs))
	}
	if envs[0] != "123" {
		t.Errorf("envs[0] = %q, want \"123\"", envs[0])
	}
}

// ---------------------------------------------------------------------------
// Tests for extractCategories
// ---------------------------------------------------------------------------

func TestExtractCategories_Normal(t *testing.T) {
	rows := [][]interface{}{
		{"Database", "Host", "localhost"},
		{"Database", "Port", "5432"},
		{"Auth", "Key", "secret"},
		{"Cache", "Driver", "redis"},
	}
	cats := extractCategories(rows)
	expected := []string{"Database", "Auth", "Cache"}
	if len(cats) != len(expected) {
		t.Fatalf("expected %d categories, got %d: %v", len(expected), len(cats), cats)
	}
	for i, e := range expected {
		if cats[i] != e {
			t.Errorf("cats[%d] = %q, want %q", i, cats[i], e)
		}
	}
}

func TestExtractCategories_EmptyRows(t *testing.T) {
	rows := [][]interface{}{}
	cats := extractCategories(rows)
	if len(cats) != 0 {
		t.Errorf("expected 0 categories for empty rows, got %d", len(cats))
	}
}

func TestExtractCategories_SkipsEmptyCategory(t *testing.T) {
	rows := [][]interface{}{
		{"Database", "Host", "localhost"},
		{"", "Port", "5432"},
		{"Auth", "Key", "secret"},
	}
	cats := extractCategories(rows)
	expected := []string{"Database", "Auth"}
	if len(cats) != len(expected) {
		t.Fatalf("expected %d categories, got %d: %v", len(expected), len(cats), cats)
	}
}

func TestExtractCategories_Deduplicated(t *testing.T) {
	rows := [][]interface{}{
		{"Database", "Host", "localhost"},
		{"Database", "Port", "5432"},
		{"Database", "Name", "mydb"},
	}
	cats := extractCategories(rows)
	if len(cats) != 1 {
		t.Errorf("expected 1 unique category, got %d: %v", len(cats), cats)
	}
	if cats[0] != "Database" {
		t.Errorf("cats[0] = %q, want \"Database\"", cats[0])
	}
}
