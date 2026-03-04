package googleapi

import (
	"testing"
)

func TestComputeDiff_Identical(t *testing.T) {
	data := map[string]interface{}{
		"__ENV__": "dev", "__SHEET__": "Env", "__SHEET_ID__": "123",
		"Database": map[string]interface{}{"Host": "localhost", "Port": "5432"},
	}
	diffs := computeDiff(data, data)
	if len(diffs) != 0 {
		t.Errorf("expected 0 diffs for identical data, got %d", len(diffs))
	}
}

func TestComputeDiff_DifferentValues(t *testing.T) {
	data1 := map[string]interface{}{
		"__ENV__": "dev", "__SHEET__": "Env", "__SHEET_ID__": "123",
		"Database": map[string]interface{}{"Host": "dev.db.com", "Port": "5432"},
	}
	data2 := map[string]interface{}{
		"__ENV__": "prod", "__SHEET__": "Env", "__SHEET_ID__": "123",
		"Database": map[string]interface{}{"Host": "prod.db.com", "Port": "5432"},
	}
	diffs := computeDiff(data1, data2)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d: %+v", len(diffs), diffs)
	}
	if diffs[0].Category != "Database" || diffs[0].Name != "Host" {
		t.Errorf("unexpected diff: %+v", diffs[0])
	}
	if diffs[0].Value1 != "dev.db.com" || diffs[0].Value2 != "prod.db.com" {
		t.Errorf("unexpected values: %q vs %q", diffs[0].Value1, diffs[0].Value2)
	}
}

func TestComputeDiff_MissingCategory(t *testing.T) {
	data1 := map[string]interface{}{
		"__ENV__": "dev", "__SHEET__": "Env", "__SHEET_ID__": "123",
		"Database": map[string]interface{}{"Host": "localhost"},
		"Cache":    map[string]interface{}{"Driver": "redis"},
	}
	data2 := map[string]interface{}{
		"__ENV__": "prod", "__SHEET__": "Env", "__SHEET_ID__": "123",
		"Database": map[string]interface{}{"Host": "localhost"},
	}
	diffs := computeDiff(data1, data2)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff for missing category, got %d: %+v", len(diffs), diffs)
	}
	if diffs[0].Category != "Cache" || diffs[0].Value2 != "" {
		t.Errorf("unexpected diff: %+v", diffs[0])
	}
}

func TestComputeDiff_MissingName(t *testing.T) {
	data1 := map[string]interface{}{
		"__ENV__": "dev", "__SHEET__": "Env", "__SHEET_ID__": "123",
		"Database": map[string]interface{}{"Host": "localhost", "Port": "5432"},
	}
	data2 := map[string]interface{}{
		"__ENV__": "prod", "__SHEET__": "Env", "__SHEET_ID__": "123",
		"Database": map[string]interface{}{"Host": "localhost"},
	}
	diffs := computeDiff(data1, data2)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff for missing name, got %d: %+v", len(diffs), diffs)
	}
	if diffs[0].Name != "Port" || diffs[0].Value1 != "5432" || diffs[0].Value2 != "" {
		t.Errorf("unexpected diff: %+v", diffs[0])
	}
}

func TestComputeDiff_AllDifferent(t *testing.T) {
	data1 := map[string]interface{}{
		"__ENV__": "dev", "__SHEET__": "Env", "__SHEET_ID__": "123",
		"A": map[string]interface{}{"X": "1"},
		"B": map[string]interface{}{"Y": "2"},
	}
	data2 := map[string]interface{}{
		"__ENV__": "prod", "__SHEET__": "Env", "__SHEET_ID__": "123",
		"A": map[string]interface{}{"X": "10"},
		"B": map[string]interface{}{"Y": "20"},
	}
	diffs := computeDiff(data1, data2)
	if len(diffs) != 2 {
		t.Errorf("expected 2 diffs, got %d: %+v", len(diffs), diffs)
	}
}

func TestAllCategoryKeys_SkipsSentinels(t *testing.T) {
	data := map[string]interface{}{
		"__ENV__": "dev", "__SHEET__": "Env", "__SHEET_ID__": "123",
		"Database": map[string]interface{}{},
	}
	keys := allCategoryKeys(data, data)
	if len(keys) != 1 || keys[0] != "Database" {
		t.Errorf("expected [Database], got %v", keys)
	}
}

func TestAllCategoryKeys_EmptyStringKey(t *testing.T) {
	data := map[string]interface{}{
		"":         "empty",
		"Database": map[string]interface{}{},
	}
	keys := allCategoryKeys(data, nil)
	if len(keys) != 1 || keys[0] != "Database" {
		t.Errorf("expected [Database], got %v", keys)
	}
}

func TestComputeDiff_MissingCategoryInFirst(t *testing.T) {
	data1 := map[string]interface{}{
		"__ENV__": "dev", "__SHEET__": "Env", "__SHEET_ID__": "123",
		"Database": map[string]interface{}{"Host": "localhost"},
	}
	data2 := map[string]interface{}{
		"__ENV__": "prod", "__SHEET__": "Env", "__SHEET_ID__": "123",
		"Database": map[string]interface{}{"Host": "localhost"},
		"Cache":    map[string]interface{}{"Driver": "redis"},
	}
	diffs := computeDiff(data1, data2)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff for category in second only, got %d: %+v", len(diffs), diffs)
	}
	if diffs[0].Category != "Cache" || diffs[0].Value1 != "" || diffs[0].Value2 != "redis" {
		t.Errorf("unexpected diff: %+v", diffs[0])
	}
}
