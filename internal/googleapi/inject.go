package googleapi

import (
	"fmt"
	"gcnf/internal/config"
	"gcnf/internal/utils"
	"os"
	"strings"
)

func InjectCommand(templatePath string, skipComments bool, configs *config.Configs) error {
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
	fmt.Println(strings.Join(outputLines, "\n"))
	return nil
}
