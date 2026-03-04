package googleapi

import (
	"fmt"
	"os"
	"strings"

	"gcnf/internal/config"
	"gcnf/internal/utils"
)

// InjectCommand processes a template file, resolving environment variables and gcnf:// URLs.
// If outputPath is non-empty, the result is written to that file; otherwise it is printed to stdout.
func InjectCommand(templatePath string, skipComments bool, outputPath string, configs *config.Configs) error {
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return err
	}
	envVars := make(map[string]string)
	lines := utils.SplitLines(string(content))
	outputLines := []string{}

	resolveGCNFURL := func(gcnfURL string) (string, error) {
		return ReadGCNFURL(gcnfURL, configs)
	}

	for _, line := range lines {
		strippedLine := utils.StripWhitespace(line)
		isComment := utils.IsCommentLine(strippedLine)
		if skipComments && isComment {
			continue
		}
		if !isComment && utils.ContainsEqualSign(strippedLine) {
			key, value := utils.SplitKeyValue(strippedLine)
			resolvedValue := utils.ResolveValue(envVars, value, resolveGCNFURL)
			envVars[key] = resolvedValue
			outputLines = append(outputLines, fmt.Sprintf("%s=%s", key, resolvedValue))
		} else {
			outputLines = append(outputLines, line)
		}
	}
	output := strings.Join(outputLines, "\n")
	if outputPath != "" {
		if err := os.WriteFile(outputPath, []byte(output+"\n"), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
	} else {
		fmt.Println(output)
	}
	return nil
}
