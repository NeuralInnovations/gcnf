package googleapi

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"gcnf/internal/config"
)

type diffEntry struct {
	Category string
	Name     string
	Value1   string
	Value2   string
}

// computeDiff compares two sheet data maps and returns entries that differ.
func computeDiff(data1, data2 map[string]interface{}) []diffEntry {
	var diffs []diffEntry
	allKeys := allCategoryKeys(data1, data2)
	for _, cat := range allKeys {
		catData1, _ := data1[cat].(map[string]interface{})
		catData2, _ := data2[cat].(map[string]interface{})
		allNames := allNameKeys(catData1, catData2)
		for _, name := range allNames {
			v1, _ := catData1[name].(string)
			v2, _ := catData2[name].(string)
			if v1 != v2 {
				diffs = append(diffs, diffEntry{cat, name, v1, v2})
			}
		}
	}
	return diffs
}

func allCategoryKeys(data1, data2 map[string]interface{}) []string {
	seen := make(map[string]bool)
	for k := range data1 {
		if k != "" && !strings.HasPrefix(k, "_") {
			seen[k] = true
		}
	}
	for k := range data2 {
		if k != "" && !strings.HasPrefix(k, "_") {
			seen[k] = true
		}
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func allNameKeys(data1, data2 map[string]interface{}) []string {
	seen := make(map[string]bool)
	for k := range data1 {
		seen[k] = true
	}
	for k := range data2 {
		seen[k] = true
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// DiffCommand loads both environments from a sheet and prints differences.
func DiffCommand(sheet, env1, env2 string, configs *config.Configs) {
	rows := loadGoogleSheet(sheet, configs)
	data1 := sheetToMap(rows, sheet, env1, configs)
	data2 := sheetToMap(rows, sheet, env2, configs)

	diffs := computeDiff(data1, data2)
	if len(diffs) == 0 {
		fmt.Println("No differences found.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "CATEGORY\tNAME\t%s\t%s\n", env1, env2)
	for _, d := range diffs {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", d.Category, d.Name, d.Value1, d.Value2)
	}
	w.Flush()
}
