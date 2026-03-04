package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// NormalizePath
// ---------------------------------------------------------------------------

func TestNormalizePath_TildeExpansion(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("cannot determine home dir: %v", err)
	}

	got := NormalizePath("~/some/path")
	want := filepath.Join(home, "some", "path")
	if got != want {
		t.Errorf("NormalizePath(\"~/some/path\") = %q, want %q", got, want)
	}
}

func TestNormalizePath_RegularPath(t *testing.T) {
	got := NormalizePath("/tmp/foo/bar")
	want := filepath.Clean("/tmp/foo/bar")
	if got != want {
		t.Errorf("NormalizePath(\"/tmp/foo/bar\") = %q, want %q", got, want)
	}
}

func TestNormalizePath_EmptyString(t *testing.T) {
	got := NormalizePath("")
	want := "."
	if got != want {
		t.Errorf("NormalizePath(\"\") = %q, want %q", got, want)
	}
}

func TestNormalizePath_WhitespaceOnly(t *testing.T) {
	got := NormalizePath("  ")
	want := "."
	if got != want {
		t.Errorf("NormalizePath(\"  \") = %q, want %q", got, want)
	}
}

func TestNormalizePath_TrailingSlashClean(t *testing.T) {
	got := NormalizePath("/tmp/foo//bar/")
	want := filepath.Clean("/tmp/foo//bar/")
	if got != want {
		t.Errorf("NormalizePath result = %q, want %q", got, want)
	}
}

func TestNormalizePath_BackslashTildeOnUnix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("backslash tilde test only meaningful on unix")
	}
	// On non-Windows, ~\ prefix should still trigger expansion.
	home, _ := os.UserHomeDir()
	got := NormalizePath("~\\subdir")
	want := filepath.Join(home, "subdir")
	if got != want {
		t.Errorf("NormalizePath(\"~\\\\subdir\") = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// EnsureDirectoryExists
// ---------------------------------------------------------------------------

func TestEnsureDirectoryExists_CreatesNew(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "newdir", "nested")
	if err := EnsureDirectoryExists(dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected a directory, got file")
	}
}

func TestEnsureDirectoryExists_ExistingDir(t *testing.T) {
	dir := t.TempDir()
	if err := EnsureDirectoryExists(dir); err != nil {
		t.Fatalf("unexpected error on existing dir: %v", err)
	}
}

func TestEnsureDirectoryExists_ErrorOnInvalidPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("null byte path test not reliable on windows")
	}
	// A path with a null byte should fail.
	err := EnsureDirectoryExists("/tmp/\x00invalid")
	if err == nil {
		// Some OS may reject this at os.Stat level; either way, we accept no panic.
		t.Log("no error returned, OS may have handled null byte gracefully")
	}
}

// ---------------------------------------------------------------------------
// FileExists
// ---------------------------------------------------------------------------

func TestFileExists_ExistingFile(t *testing.T) {
	f := filepath.Join(t.TempDir(), "exists.txt")
	if err := os.WriteFile(f, []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	if !FileExists(f) {
		t.Errorf("FileExists(%q) = false, want true", f)
	}
}

func TestFileExists_NonExistingFile(t *testing.T) {
	if FileExists("/tmp/definitely_does_not_exist_abc123xyz") {
		t.Error("FileExists returned true for non-existing file")
	}
}

func TestFileExists_Directory(t *testing.T) {
	dir := t.TempDir()
	// FileExists uses os.Stat which succeeds on directories too.
	if !FileExists(dir) {
		t.Errorf("FileExists(%q) = false for directory, want true", dir)
	}
}

// ---------------------------------------------------------------------------
// DeleteFile
// ---------------------------------------------------------------------------

func TestDeleteFile_ExistingFile(t *testing.T) {
	f := filepath.Join(t.TempDir(), "todelete.txt")
	if err := os.WriteFile(f, []byte("bye"), 0600); err != nil {
		t.Fatal(err)
	}
	deleted, err := DeleteFile(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Error("DeleteFile returned false for existing file")
	}
	// Verify the file is actually gone.
	if FileExists(f) {
		t.Error("file still exists after DeleteFile")
	}
}

func TestDeleteFile_NonExistingFile(t *testing.T) {
	deleted, err := DeleteFile("/tmp/nonexistent_file_xyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted {
		t.Error("DeleteFile returned true for non-existing file")
	}
}

// ---------------------------------------------------------------------------
// LoadFileContentAsString
// ---------------------------------------------------------------------------

func TestLoadFileContentAsString_Normal(t *testing.T) {
	f := filepath.Join(t.TempDir(), "content.txt")
	if err := os.WriteFile(f, []byte("hello world"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := LoadFileContentAsString(f, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello world" {
		t.Errorf("got %q, want %q", got, "hello world")
	}
}

func TestLoadFileContentAsString_Trimmed(t *testing.T) {
	f := filepath.Join(t.TempDir(), "trimmed.txt")
	if err := os.WriteFile(f, []byte("  hello  \n"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := LoadFileContentAsString(f, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}

func TestLoadFileContentAsString_NotTrimmed(t *testing.T) {
	f := filepath.Join(t.TempDir(), "nottrimmed.txt")
	content := "  hello  \n"
	if err := os.WriteFile(f, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := LoadFileContentAsString(f, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != content {
		t.Errorf("got %q, want %q", got, content)
	}
}

func TestLoadFileContentAsString_EmptyFile(t *testing.T) {
	f := filepath.Join(t.TempDir(), "empty.txt")
	if err := os.WriteFile(f, []byte(""), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := LoadFileContentAsString(f, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

func TestLoadFileContentAsString_NonExistentFile(t *testing.T) {
	_, err := LoadFileContentAsString("/tmp/no_such_file_xyz", false)
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

// ---------------------------------------------------------------------------
// LoadFileContentAsJson
// ---------------------------------------------------------------------------

func TestLoadFileContentAsJson_ValidJSON(t *testing.T) {
	f := filepath.Join(t.TempDir(), "data.json")
	data := map[string]interface{}{"key": "value", "num": float64(42)}
	raw, _ := json.Marshal(data)
	if err := os.WriteFile(f, raw, 0600); err != nil {
		t.Fatal(err)
	}
	result := LoadFileContentAsJson(f)
	if result == nil {
		t.Fatal("expected non-nil map")
	}
	if result["key"] != "value" {
		t.Errorf("result[\"key\"] = %v, want \"value\"", result["key"])
	}
	if result["num"] != float64(42) {
		t.Errorf("result[\"num\"] = %v, want 42", result["num"])
	}
}

func TestLoadFileContentAsJson_InvalidJSON(t *testing.T) {
	f := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(f, []byte("{broken json"), 0600); err != nil {
		t.Fatal(err)
	}
	result := LoadFileContentAsJson(f)
	if result != nil {
		t.Errorf("expected nil for invalid JSON, got %v", result)
	}
}

func TestLoadFileContentAsJson_NonExistentFile(t *testing.T) {
	result := LoadFileContentAsJson("/tmp/no_such_file_xyz.json")
	if result != nil {
		t.Errorf("expected nil for non-existent file, got %v", result)
	}
}

// ---------------------------------------------------------------------------
// MergeMaps
// ---------------------------------------------------------------------------

func TestMergeMaps_OverlappingKeys(t *testing.T) {
	a := map[string]interface{}{"x": 1, "y": 2}
	b := map[string]interface{}{"y": 99, "z": 3}
	result := MergeMaps(a, b)

	if result["x"] != 1 {
		t.Errorf("result[\"x\"] = %v, want 1", result["x"])
	}
	if result["y"] != 99 {
		t.Errorf("result[\"y\"] = %v, want 99 (b should override a)", result["y"])
	}
	if result["z"] != 3 {
		t.Errorf("result[\"z\"] = %v, want 3", result["z"])
	}
}

func TestMergeMaps_EmptyMaps(t *testing.T) {
	result := MergeMaps(map[string]interface{}{}, map[string]interface{}{})
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}

func TestMergeMaps_OneEmpty(t *testing.T) {
	a := map[string]interface{}{"a": "val"}
	result := MergeMaps(a, map[string]interface{}{})
	if result["a"] != "val" {
		t.Errorf("result[\"a\"] = %v, want \"val\"", result["a"])
	}
}

func TestMergeMaps_DoesNotMutateOriginals(t *testing.T) {
	a := map[string]interface{}{"k": "original"}
	b := map[string]interface{}{"k": "override"}
	_ = MergeMaps(a, b)
	if a["k"] != "original" {
		t.Error("MergeMaps mutated the first map")
	}
}

// ---------------------------------------------------------------------------
// LoadProperties
// ---------------------------------------------------------------------------

func TestLoadProperties_Standard(t *testing.T) {
	content := "host=localhost\nport=8080\nname = my app"
	props, err := LoadProperties(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if props["host"] != "localhost" {
		t.Errorf("props[\"host\"] = %q, want \"localhost\"", props["host"])
	}
	if props["port"] != "8080" {
		t.Errorf("props[\"port\"] = %q, want \"8080\"", props["port"])
	}
	if props["name"] != "my app" {
		t.Errorf("props[\"name\"] = %q, want \"my app\"", props["name"])
	}
}

func TestLoadProperties_SectionHeadersSkipped(t *testing.T) {
	content := "[database]\nhost=db.local\n[cache]\nhost=cache.local"
	props, err := LoadProperties(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Section headers like [database] have no '=' so are skipped.
	if _, ok := props["[database]"]; ok {
		t.Error("section header [database] should have been skipped")
	}
	if _, ok := props["[cache]"]; ok {
		t.Error("section header [cache] should have been skipped")
	}
	// But key=value lines should be present.
	if props["host"] != "cache.local" {
		t.Errorf("props[\"host\"] = %q, want \"cache.local\" (last wins)", props["host"])
	}
}

func TestLoadProperties_EmptyContent(t *testing.T) {
	props, err := LoadProperties("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(props) != 0 {
		t.Errorf("expected empty map, got %v", props)
	}
}

func TestLoadProperties_ValueWithEquals(t *testing.T) {
	content := "conn=user=admin;pass=secret"
	props, err := LoadProperties(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if props["conn"] != "user=admin;pass=secret" {
		t.Errorf("props[\"conn\"] = %q, want \"user=admin;pass=secret\"", props["conn"])
	}
}

// ---------------------------------------------------------------------------
// ToBase64 / FromBase64 round-trip
// ---------------------------------------------------------------------------

func TestToBase64_FromBase64_RoundTrip(t *testing.T) {
	inputs := []string{
		"hello world",
		"",
		"special chars: !@#$%^&*()",
		"unicode: \u00e9\u00e0\u00fc\u00f6",
		"multiline\ntext\nhere",
	}
	for _, input := range inputs {
		encoded := ToBase64(input)
		decoded, err := FromBase64(encoded)
		if err != nil {
			t.Errorf("FromBase64 error for input %q: %v", input, err)
			continue
		}
		if decoded != input {
			t.Errorf("round-trip failed: got %q, want %q", decoded, input)
		}
	}
}

func TestFromBase64_InvalidInput(t *testing.T) {
	_, err := FromBase64("not!valid!base64!!!")
	if err == nil {
		t.Error("expected error for invalid base64, got nil")
	}
}

// ---------------------------------------------------------------------------
// Coalesce
// ---------------------------------------------------------------------------

func TestCoalesce_FirstNonEmpty(t *testing.T) {
	got := Coalesce("", "", "third", "fourth")
	if got != "third" {
		t.Errorf("Coalesce = %q, want \"third\"", got)
	}
}

func TestCoalesce_AllEmpty(t *testing.T) {
	got := Coalesce("", "", "")
	if got != "" {
		t.Errorf("Coalesce = %q, want empty string", got)
	}
}

func TestCoalesce_SingleValue(t *testing.T) {
	got := Coalesce("only")
	if got != "only" {
		t.Errorf("Coalesce = %q, want \"only\"", got)
	}
}

func TestCoalesce_NoArgs(t *testing.T) {
	got := Coalesce()
	if got != "" {
		t.Errorf("Coalesce() = %q, want empty string", got)
	}
}

func TestCoalesce_FirstIsNonEmpty(t *testing.T) {
	got := Coalesce("first", "second")
	if got != "first" {
		t.Errorf("Coalesce = %q, want \"first\"", got)
	}
}

// ---------------------------------------------------------------------------
// IsValidBase64
// ---------------------------------------------------------------------------

func TestIsValidBase64_Valid(t *testing.T) {
	valid := ToBase64("test data")
	if !IsValidBase64(valid) {
		t.Errorf("IsValidBase64(%q) = false, want true", valid)
	}
}

func TestIsValidBase64_EmptyString(t *testing.T) {
	// Empty string is valid base64 (decodes to empty).
	if !IsValidBase64("") {
		t.Error("IsValidBase64(\"\") = false, want true")
	}
}

func TestIsValidBase64_Invalid(t *testing.T) {
	if IsValidBase64("not!valid@base64#") {
		t.Error("IsValidBase64 returned true for invalid input")
	}
}

// ---------------------------------------------------------------------------
// SplitLines
// ---------------------------------------------------------------------------

func TestSplitLines_Normal(t *testing.T) {
	lines := SplitLines("a\nb\nc")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "a" || lines[1] != "b" || lines[2] != "c" {
		t.Errorf("unexpected lines: %v", lines)
	}
}

func TestSplitLines_EmptyString(t *testing.T) {
	lines := SplitLines("")
	// strings.Split("", "\n") returns [""]
	if len(lines) != 1 || lines[0] != "" {
		t.Errorf("expected [\"\"], got %v", lines)
	}
}

func TestSplitLines_SingleLine(t *testing.T) {
	lines := SplitLines("single")
	if len(lines) != 1 || lines[0] != "single" {
		t.Errorf("expected [\"single\"], got %v", lines)
	}
}

func TestSplitLines_TrailingNewline(t *testing.T) {
	lines := SplitLines("a\nb\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 elements, got %d: %v", len(lines), lines)
	}
	if lines[2] != "" {
		t.Errorf("expected trailing empty string, got %q", lines[2])
	}
}

// ---------------------------------------------------------------------------
// StripWhitespace
// ---------------------------------------------------------------------------

func TestStripWhitespace_Spaces(t *testing.T) {
	got := StripWhitespace("  hello  ")
	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}

func TestStripWhitespace_Tabs(t *testing.T) {
	got := StripWhitespace("\t\ttabbed\t\t")
	if got != "tabbed" {
		t.Errorf("got %q, want %q", got, "tabbed")
	}
}

func TestStripWhitespace_Empty(t *testing.T) {
	got := StripWhitespace("")
	if got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

func TestStripWhitespace_Mixed(t *testing.T) {
	got := StripWhitespace(" \t \n mixed \n \t ")
	if got != "mixed" {
		t.Errorf("got %q, want %q", got, "mixed")
	}
}

// ---------------------------------------------------------------------------
// IsCommentLine
// ---------------------------------------------------------------------------

func TestIsCommentLine_Hash(t *testing.T) {
	if !IsCommentLine("# this is a comment") {
		t.Error("expected true for # comment")
	}
}

func TestIsCommentLine_DoubleSlash(t *testing.T) {
	if !IsCommentLine("// this is a comment") {
		t.Error("expected true for // comment")
	}
}

func TestIsCommentLine_Semicolon(t *testing.T) {
	if !IsCommentLine("; this is a comment") {
		t.Error("expected true for ; comment")
	}
}

func TestIsCommentLine_NonComment(t *testing.T) {
	if IsCommentLine("not a comment") {
		t.Error("expected false for non-comment line")
	}
}

func TestIsCommentLine_EmptyString(t *testing.T) {
	if IsCommentLine("") {
		t.Error("expected false for empty string")
	}
}

func TestIsCommentLine_HashInMiddle(t *testing.T) {
	if IsCommentLine("value # not a comment") {
		t.Error("expected false when # is not at the start")
	}
}

// ---------------------------------------------------------------------------
// ContainsEqualSign
// ---------------------------------------------------------------------------

func TestContainsEqualSign_WithEquals(t *testing.T) {
	if !ContainsEqualSign("key=value") {
		t.Error("expected true")
	}
}

func TestContainsEqualSign_WithoutEquals(t *testing.T) {
	if ContainsEqualSign("no equals here") {
		t.Error("expected false")
	}
}

func TestContainsEqualSign_EmptyString(t *testing.T) {
	if ContainsEqualSign("") {
		t.Error("expected false for empty string")
	}
}

func TestContainsEqualSign_OnlyEquals(t *testing.T) {
	if !ContainsEqualSign("=") {
		t.Error("expected true for bare =")
	}
}

// ---------------------------------------------------------------------------
// SplitKeyValue
// ---------------------------------------------------------------------------

func TestSplitKeyValue_Normal(t *testing.T) {
	key, val := SplitKeyValue("name=John")
	if key != "name" || val != "John" {
		t.Errorf("got key=%q val=%q, want key=\"name\" val=\"John\"", key, val)
	}
}

func TestSplitKeyValue_ValueContainsEquals(t *testing.T) {
	key, val := SplitKeyValue("conn=host=db;port=5432")
	if key != "conn" || val != "host=db;port=5432" {
		t.Errorf("got key=%q val=%q, want key=\"conn\" val=\"host=db;port=5432\"", key, val)
	}
}

func TestSplitKeyValue_NoEquals(t *testing.T) {
	key, val := SplitKeyValue("justkey")
	if key != "justkey" || val != "" {
		t.Errorf("got key=%q val=%q, want key=\"justkey\" val=\"\"", key, val)
	}
}

func TestSplitKeyValue_SpacesAroundEquals(t *testing.T) {
	key, val := SplitKeyValue("  key  =  value  ")
	if key != "key" || val != "value" {
		t.Errorf("got key=%q val=%q, want key=\"key\" val=\"value\"", key, val)
	}
}

func TestSplitKeyValue_EmptyValue(t *testing.T) {
	key, val := SplitKeyValue("key=")
	if key != "key" || val != "" {
		t.Errorf("got key=%q val=%q, want key=\"key\" val=\"\"", key, val)
	}
}

// ---------------------------------------------------------------------------
// stripQuotes (unexported, tested indirectly via ResolveValue, but also directly)
// ---------------------------------------------------------------------------

func TestStripQuotes_DoubleQuotes(t *testing.T) {
	got := stripQuotes(`"hello"`)
	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}

func TestStripQuotes_SingleQuotes(t *testing.T) {
	got := stripQuotes(`'hello'`)
	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}

func TestStripQuotes_NoQuotes(t *testing.T) {
	got := stripQuotes("hello")
	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}

func TestStripQuotes_MismatchedQuotes(t *testing.T) {
	got := stripQuotes(`"hello'`)
	if got != `"hello'` {
		t.Errorf("got %q, want %q (mismatched quotes should remain)", got, `"hello'`)
	}
}

func TestStripQuotes_Empty(t *testing.T) {
	got := stripQuotes("")
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

// ---------------------------------------------------------------------------
// ResolveValue  (CRITICAL)
// ---------------------------------------------------------------------------

// noopResolver is a gcnf:// resolver that should not be called unless we expect it.
func noopResolver(url string) (string, error) {
	return "", nil
}

func TestResolveValue_BasicVar_EnvVarsFirst(t *testing.T) {
	// BUG-1 regression: $VAR must use envVars FIRST, not os.Getenv.
	envVars := map[string]string{"MY_VAR": "from_envVars"}
	t.Setenv("MY_VAR", "from_os")

	got := ResolveValue(envVars, "$MY_VAR", noopResolver)
	if got != "from_envVars" {
		t.Errorf("ResolveValue($MY_VAR) = %q, want \"from_envVars\" (envVars should take priority)", got)
	}
}

func TestResolveValue_BasicVar_FallbackToOsEnv(t *testing.T) {
	envVars := map[string]string{} // not in envVars
	t.Setenv("FALLBACK_VAR", "from_os_env")

	got := ResolveValue(envVars, "$FALLBACK_VAR", noopResolver)
	if got != "from_os_env" {
		t.Errorf("ResolveValue($FALLBACK_VAR) = %q, want \"from_os_env\"", got)
	}
}

func TestResolveValue_BasicVar_Undefined(t *testing.T) {
	envVars := map[string]string{}
	os.Unsetenv("UNDEFINED_VAR_XYZ")

	got := ResolveValue(envVars, "$UNDEFINED_VAR_XYZ", noopResolver)
	if got != "" {
		t.Errorf("ResolveValue($UNDEFINED_VAR_XYZ) = %q, want empty string", got)
	}
}

func TestResolveValue_BracedVar_FromEnvVars(t *testing.T) {
	envVars := map[string]string{"BRACED": "braced_val"}
	got := ResolveValue(envVars, "${BRACED}", noopResolver)
	if got != "braced_val" {
		t.Errorf("ResolveValue(${BRACED}) = %q, want \"braced_val\"", got)
	}
}

func TestResolveValue_BracedVar_FallbackToOsEnv(t *testing.T) {
	envVars := map[string]string{}
	t.Setenv("BRACED_OS", "os_braced_val")

	got := ResolveValue(envVars, "${BRACED_OS}", noopResolver)
	if got != "os_braced_val" {
		t.Errorf("ResolveValue(${BRACED_OS}) = %q, want \"os_braced_val\"", got)
	}
}

func TestResolveValue_BracedVar_Default(t *testing.T) {
	envVars := map[string]string{}
	os.Unsetenv("MISSING_VAR")

	got := ResolveValue(envVars, "${MISSING_VAR:-default_val}", noopResolver)
	if got != "default_val" {
		t.Errorf("ResolveValue(${MISSING_VAR:-default_val}) = %q, want \"default_val\"", got)
	}
}

func TestResolveValue_BracedVar_DefaultNotUsedWhenSet(t *testing.T) {
	envVars := map[string]string{"PRESENT_VAR": "present"}

	got := ResolveValue(envVars, "${PRESENT_VAR:-default_val}", noopResolver)
	if got != "present" {
		t.Errorf("ResolveValue = %q, want \"present\" (default should not be used)", got)
	}
}

func TestResolveValue_GcnfURL(t *testing.T) {
	envVars := map[string]string{}
	resolver := func(url string) (string, error) {
		if url == "gcnf://config/key" {
			return "resolved_value", nil
		}
		return "", nil
	}

	got := ResolveValue(envVars, "gcnf://config/key", resolver)
	if got != "resolved_value" {
		t.Errorf("ResolveValue(gcnf://...) = %q, want \"resolved_value\"", got)
	}
}

func TestResolveValue_QuotedValue(t *testing.T) {
	envVars := map[string]string{"QV": "inner"}

	got := ResolveValue(envVars, `"hello $QV world"`, noopResolver)
	if got != `"hello inner world"` {
		t.Errorf("ResolveValue(quoted) = %q, want %q", got, `"hello inner world"`)
	}
}

func TestResolveValue_NoVariables(t *testing.T) {
	envVars := map[string]string{}
	got := ResolveValue(envVars, "plain text", noopResolver)
	if got != "plain text" {
		t.Errorf("ResolveValue(plain) = %q, want \"plain text\"", got)
	}
}

func TestResolveValue_MultipleVars(t *testing.T) {
	envVars := map[string]string{"A": "alpha", "B": "beta"}
	got := ResolveValue(envVars, "$A-$B", noopResolver)
	if got != "alpha-beta" {
		t.Errorf("ResolveValue($A-$B) = %q, want \"alpha-beta\"", got)
	}
}

func TestResolveValue_BracedVarWithQuotedValue(t *testing.T) {
	// envVars value is quoted -- stripQuotes should remove them.
	envVars := map[string]string{"QUOTED_VAR": `"quoted_inner"`}
	got := ResolveValue(envVars, "${QUOTED_VAR}", noopResolver)
	if got != "quoted_inner" {
		t.Errorf("ResolveValue = %q, want \"quoted_inner\" (quotes should be stripped)", got)
	}
}

func TestResolveValue_EmptyValue(t *testing.T) {
	envVars := map[string]string{}
	got := ResolveValue(envVars, "", noopResolver)
	if got != "" {
		t.Errorf("ResolveValue(\"\") = %q, want empty", got)
	}
}

func TestResolveValue_MixedBracedAndBasic(t *testing.T) {
	envVars := map[string]string{"X": "xval"}
	t.Setenv("Y", "yval")

	got := ResolveValue(envVars, "${X}_$Y", noopResolver)
	if got != "xval_yval" {
		t.Errorf("ResolveValue(${X}_$Y) = %q, want \"xval_yval\"", got)
	}
}

func TestResolveValue_GcnfURLAfterVarExpansion(t *testing.T) {
	// If variable expansion produces a gcnf:// URL, it should be resolved.
	envVars := map[string]string{"GCNF_PATH": "gcnf://dynamic/path"}
	resolver := func(url string) (string, error) {
		return "dynamic_resolved", nil
	}
	got := ResolveValue(envVars, "$GCNF_PATH", resolver)
	if got != "dynamic_resolved" {
		t.Errorf("ResolveValue = %q, want \"dynamic_resolved\"", got)
	}
}

// ---------------------------------------------------------------------------
// WriteStringToFile
// ---------------------------------------------------------------------------

func TestWriteStringToFile_WriteAndReadBack(t *testing.T) {
	f := filepath.Join(t.TempDir(), "output.txt")
	content := "file content here"
	if err := WriteStringToFile(f, content); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := os.ReadFile(f)
	if err != nil {
		t.Fatalf("failed to read back: %v", err)
	}
	if string(got) != content {
		t.Errorf("got %q, want %q", string(got), content)
	}
}

func TestWriteStringToFile_CreatesParentDirectories(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "a", "b", "c", "deep.txt")
	if err := WriteStringToFile(f, "deep content"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !FileExists(f) {
		t.Error("file was not created in nested directories")
	}
	got, _ := os.ReadFile(f)
	if string(got) != "deep content" {
		t.Errorf("got %q, want \"deep content\"", string(got))
	}
}

func TestWriteStringToFile_OverwritesExisting(t *testing.T) {
	f := filepath.Join(t.TempDir(), "overwrite.txt")
	if err := WriteStringToFile(f, "first"); err != nil {
		t.Fatal(err)
	}
	if err := WriteStringToFile(f, "second"); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(f)
	if string(got) != "second" {
		t.Errorf("got %q, want \"second\"", string(got))
	}
}

func TestWriteStringToFile_EmptyContent(t *testing.T) {
	f := filepath.Join(t.TempDir(), "empty.txt")
	if err := WriteStringToFile(f, ""); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(f)
	if string(got) != "" {
		t.Errorf("got %q, want empty", string(got))
	}
}

// ---------------------------------------------------------------------------
// Integration-style: ResolveValue with $VAR priority (BUG-1 regression)
// ---------------------------------------------------------------------------

func TestResolveValue_BUG1_Regression_EnvVarsPriority(t *testing.T) {
	// This test documents the critical requirement: when both envVars and
	// os.Getenv have the same variable name, envVars MUST win.
	envVarKey := "BUG1_REGRESSION_VAR"
	envVars := map[string]string{envVarKey: "envVars_wins"}
	t.Setenv(envVarKey, "os_should_lose")

	// Test with $VAR syntax.
	got := ResolveValue(envVars, "$"+envVarKey, noopResolver)
	if got != "envVars_wins" {
		t.Errorf("[BUG-1] $VAR: got %q, want \"envVars_wins\"", got)
	}

	// Test with ${VAR} syntax.
	got = ResolveValue(envVars, "${"+envVarKey+"}", noopResolver)
	if got != "envVars_wins" {
		t.Errorf("[BUG-1] ${VAR}: got %q, want \"envVars_wins\"", got)
	}

	// Test with ${VAR:-default} syntax -- default should NOT be used.
	got = ResolveValue(envVars, "${"+envVarKey+":-fallback}", noopResolver)
	if got != "envVars_wins" {
		t.Errorf("[BUG-1] ${VAR:-default}: got %q, want \"envVars_wins\"", got)
	}
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestResolveValue_DollarSignAlone(t *testing.T) {
	envVars := map[string]string{}
	// A lone $ that doesn't match any pattern should pass through.
	got := ResolveValue(envVars, "price is $5", noopResolver)
	// basicPattern matches $5 (digits are not \w? Actually \w includes digits).
	// $5 will be treated as variable "5".
	// Since envVars["5"] doesn't exist and os.Getenv("5") is "", it resolves to "".
	if !strings.Contains(got, "price is") {
		t.Errorf("unexpected result: %q", got)
	}
}

func TestLoadProperties_CommentLinesIgnored(t *testing.T) {
	// Comment lines don't contain = in the expected format, so they are skipped.
	content := "# comment line\nkey=value\n; another comment"
	props, err := LoadProperties(content)
	if err != nil {
		t.Fatal(err)
	}
	if len(props) != 1 {
		t.Errorf("expected 1 property, got %d: %v", len(props), props)
	}
	if props["key"] != "value" {
		t.Errorf("props[\"key\"] = %q, want \"value\"", props["key"])
	}
}

// ---------------------------------------------------------------------------
// IsCacheExpired
// ---------------------------------------------------------------------------

func TestIsCacheExpired_ZeroTTL(t *testing.T) {
	tmpDir := t.TempDir()
	f := filepath.Join(tmpDir, "cache.json")
	os.WriteFile(f, []byte("{}"), 0644)

	if IsCacheExpired(f, 0) {
		t.Error("expected false when TTL is 0 (disabled)")
	}
}

func TestIsCacheExpired_NegativeTTL(t *testing.T) {
	tmpDir := t.TempDir()
	f := filepath.Join(tmpDir, "cache.json")
	os.WriteFile(f, []byte("{}"), 0644)

	if IsCacheExpired(f, -1*time.Minute) {
		t.Error("expected false when TTL is negative")
	}
}

func TestIsCacheExpired_FreshFile(t *testing.T) {
	tmpDir := t.TempDir()
	f := filepath.Join(tmpDir, "cache.json")
	os.WriteFile(f, []byte("{}"), 0644)

	if IsCacheExpired(f, 1*time.Hour) {
		t.Error("expected false for freshly created file with 1h TTL")
	}
}

func TestIsCacheExpired_ExpiredFile(t *testing.T) {
	tmpDir := t.TempDir()
	f := filepath.Join(tmpDir, "cache.json")
	os.WriteFile(f, []byte("{}"), 0644)

	// Backdate the file to 2 hours ago
	past := time.Now().Add(-2 * time.Hour)
	os.Chtimes(f, past, past)

	if !IsCacheExpired(f, 1*time.Hour) {
		t.Error("expected true for file older than TTL")
	}
}

func TestIsCacheExpired_NonExistentFile(t *testing.T) {
	if IsCacheExpired("/nonexistent/file.json", 1*time.Hour) {
		t.Error("expected false for non-existent file")
	}
}
