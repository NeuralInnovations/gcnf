package googleapi

import (
	"fmt"
	"os"
	"strings"

	"gcnf/internal/config"
	"gcnf/internal/utils"
)

// ValidationIssue represents a single issue found during template validation.
type ValidationIssue struct {
	Line   int
	Type   string // "missing_var", "malformed_url", "empty_value"
	Detail string
}

func (v ValidationIssue) String() string {
	return fmt.Sprintf("line %d: [%s] %s", v.Line, v.Type, v.Detail)
}

// ValidateCommand parses a template file and returns validation issues
// without actually resolving values.
func ValidateCommand(templatePath string, configs *config.Configs) []ValidationIssue {
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return []ValidationIssue{{Line: 0, Type: "error", Detail: err.Error()}}
	}

	var issues []ValidationIssue
	envVars := make(map[string]string)
	lines := utils.SplitLines(string(content))

	for i, line := range lines {
		lineNum := i + 1
		strippedLine := utils.StripWhitespace(line)
		if utils.IsCommentLine(strippedLine) || strippedLine == "" {
			continue
		}
		if !utils.ContainsEqualSign(strippedLine) {
			continue
		}

		key, value := utils.SplitKeyValue(strippedLine)

		// Check braced variables: ${VAR}, ${VAR:-default}, ${VAR:?error}
		for _, match := range utils.BracedPattern.FindAllStringSubmatch(value, -1) {
			varName := match[1]
			defaultVal := match[2]
			if _, ok := envVars[varName]; !ok && os.Getenv(varName) == "" && defaultVal == "" {
				issues = append(issues, ValidationIssue{lineNum, "missing_var", fmt.Sprintf("${%s} is undefined and has no default", varName)})
			}
		}

		// Check basic variables: $VAR (after removing braced patterns to avoid double-matching)
		cleanValue := utils.BracedPattern.ReplaceAllString(value, "")
		for _, match := range utils.BasicPattern.FindAllStringSubmatch(cleanValue, -1) {
			varName := match[1]
			if _, ok := envVars[varName]; !ok && os.Getenv(varName) == "" {
				issues = append(issues, ValidationIssue{lineNum, "missing_var", fmt.Sprintf("$%s is undefined", varName)})
			}
		}

		// Check gcnf:// URL format
		if strings.HasPrefix(value, "gcnf://") {
			parts := strings.SplitN(value[7:], "/", 4)
			if len(parts) != 4 {
				issues = append(issues, ValidationIssue{lineNum, "malformed_url", fmt.Sprintf("gcnf:// URL should have format gcnf://SHEET/ENV/CATEGORY/NAME, got: %s", value)})
			}
		}

		// Check empty value
		if value == "" {
			issues = append(issues, ValidationIssue{lineNum, "empty_value", fmt.Sprintf("%s has an empty value", key)})
		}

		// Track defined variables for subsequent lines
		envVars[key] = value
	}
	return issues
}
