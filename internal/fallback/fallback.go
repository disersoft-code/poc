package fallback

import (
	"errors"
	"fmt"
	"strings"

	"codeAct-poc/internal/model"
)

func GeneratePlan(task string, language string) (model.Plan, error) {
	normalizedTask := strings.TrimSpace(task)
	lowerTask := strings.ToLower(normalizedTask)

	intent := extractIntent(lowerTask)
	if intent == "" {
		return model.Plan{}, errors.New("could not identify the requested intent")
	}

	target := extractTarget(normalizedTask)
	if target == "" {
		return model.Plan{}, errors.New("could not identify the target path")
	}

	patterns := extractPatterns(lowerTask)

	script, err := buildScript(language, intent, target, patterns)
	if err != nil {
		return model.Plan{}, err
	}

	return model.Plan{
		UseCase:  "log_agent",
		Intent:   intent,
		Language: language,
		Target:   target,
		Patterns: patterns,
		Script:   script,
	}, nil
}

func extractIntent(task string) string {
	switch {
	case strings.Contains(task, "count"):
		return "count_matches"
	case strings.Contains(task, "list"):
		return "list_matches"
	default:
		return ""
	}
}

func extractPatterns(task string) []string {
	patterns := make([]string, 0)

	if strings.Contains(task, "error") {
		patterns = append(patterns, "error")
	}

	if strings.Contains(task, "warning") {
		patterns = append(patterns, "warning")
	}

	if len(patterns) == 0 {
		patterns = append(patterns, "error")
	}

	return patterns
}

func extractTarget(task string) string {
	lowerTask := strings.ToLower(task)

	marker := " in "
	index := strings.LastIndex(lowerTask, marker)
	if index == -1 {
		return ""
	}

	target := strings.TrimSpace(task[index+len(marker):])
	target = strings.Trim(target, `"'`)

	return target
}

func buildScript(language string, intent string, target string, patterns []string) (string, error) {
	switch language {
	case "powershell":
		return buildPowerShellScript(intent, target, patterns), nil
	case "bash":
		return buildBashScript(intent, target, patterns), nil
	default:
		return "", fmt.Errorf("unsupported script language: %s", language)
	}
}

func buildPowerShellScript(intent string, target string, patterns []string) string {
	condition := buildPowerShellCondition(patterns)

	switch intent {
	case "count_matches":
		return fmt.Sprintf(`$files = Get-ChildItem -Path '%s' -File
$count = 0

foreach ($file in $files) {
    $lineNumber = 0
    Get-Content $file.FullName | ForEach-Object {
        $lineNumber++
        $lower = $_.ToLower()
        if (%s) {
            $count++
        }
    }
}

Write-Output $count
`, target, condition)

	case "list_matches":
		return fmt.Sprintf(`$files = Get-ChildItem -Path '%s' -File

foreach ($file in $files) {
    $lineNumber = 0
    Get-Content $file.FullName | ForEach-Object {
        $lineNumber++
        $lower = $_.ToLower()
        if (%s) {
            Write-Output "$($file.FullName):${lineNumber}: $_"
        }
    }
}
`, target, condition)

	default:
		return ""
	}
}

func buildBashScript(intent string, target string, patterns []string) string {
	condition := buildBashCondition(patterns)

	switch intent {
	case "count_matches":
		return fmt.Sprintf(`count=0

while IFS= read -r file; do
  while IFS= read -r line; do
    lower="$(printf '%%s' "$line" | tr '[:upper:]' '[:lower:]')"
    if %s; then
      count=$((count + 1))
    fi
  done < "$file"
done < <(find '%s' -maxdepth 1 -type f)

printf '%%s\n' "$count"
`, condition, target)

	case "list_matches":
		return fmt.Sprintf(`while IFS= read -r file; do
  line_number=0
  while IFS= read -r line; do
    line_number=$((line_number + 1))
    lower="$(printf '%%s' "$line" | tr '[:upper:]' '[:lower:]')"
    if %s; then
      printf '%%s:%%s: %%s\n' "$file" "$line_number" "$line"
    fi
  done < "$file"
done < <(find '%s' -maxdepth 1 -type f)
`, condition, target)

	default:
		return ""
	}
}

func buildPowerShellCondition(patterns []string) string {
	parts := make([]string, 0, len(patterns))

	for _, pattern := range patterns {
		parts = append(parts, fmt.Sprintf(`$lower.Contains('%s')`, pattern))
	}

	return strings.Join(parts, " -or ")
}

func buildBashCondition(patterns []string) string {
	parts := make([]string, 0, len(patterns))

	for _, pattern := range patterns {
		parts = append(parts, fmt.Sprintf(`[[ "$lower" == *%s* ]]`, pattern))
	}

	return strings.Join(parts, " || ")
}
