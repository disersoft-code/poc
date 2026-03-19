package formatter

import (
	"fmt"
	"strconv"
	"strings"

	"codeAct-poc/internal/model"
)

func Format(plan model.Plan, result model.ExecutionResult) string {
	switch plan.Intent {
	case "count_matches":
		return formatCountResult(plan, result)
	case "list_matches", "find_text":
		return formatListResult(plan, result)
	case "create_file":
		return formatSimpleOperationResult(result, "The file was created successfully.")
	case "delete_path":
		return formatSimpleOperationResult(result, "The path was deleted successfully.")
	case "rename_path":
		return formatSimpleOperationResult(result, "The path was renamed successfully.")
	case "copy_path":
		return formatSimpleOperationResult(result, "The path was copied successfully.")
	case "replace_text":
		return formatSimpleOperationResult(result, "The text was replaced successfully.")
	case "count_files":
		return formatCountFilesResult(plan, result)
	case "list_files":
		return formatListFilesResult(plan, result)
	default:
		return formatDefaultResult(result)
	}
}

func FormatDebugPlan(plan model.Plan) string {
	builder := strings.Builder{}

	builder.WriteString("---- PLAN DEBUG ----\n")
	builder.WriteString(fmt.Sprintf("UseCase: %s\n", plan.UseCase))
	builder.WriteString(fmt.Sprintf("Intent: %s\n", plan.Intent))
	builder.WriteString(fmt.Sprintf("Language: %s\n", plan.Language))
	builder.WriteString(fmt.Sprintf("Target: %s\n", plan.Target))

	if strings.TrimSpace(plan.Destination) != "" {
		builder.WriteString(fmt.Sprintf("Destination: %s\n", plan.Destination))
	}

	if len(plan.Patterns) > 0 {
		builder.WriteString(fmt.Sprintf("Patterns: %v\n", plan.Patterns))
	} else {
		builder.WriteString("Patterns: []\n")
	}

	if strings.TrimSpace(plan.Replacement) != "" {
		builder.WriteString(fmt.Sprintf("Replacement: %s\n", plan.Replacement))
	}

	if strings.TrimSpace(plan.Content) != "" {
		builder.WriteString(fmt.Sprintf("Content: %s\n", plan.Content))
	}

	builder.WriteString(fmt.Sprintf("Source: %s\n", plan.Source))
	builder.WriteString("Script:\n")
	builder.WriteString(plan.Script)
	builder.WriteString("\n--------------------")

	return builder.String()
}

func FormatDebugResult(result model.ExecutionResult, err error) string {
	builder := strings.Builder{}

	builder.WriteString("---- EXECUTION DEBUG ----\n")
	builder.WriteString(fmt.Sprintf("ExitCode: %d\n", result.ExitCode))

	if strings.TrimSpace(result.Stdout) != "" {
		builder.WriteString("Stdout:\n")
		builder.WriteString(result.Stdout)
		builder.WriteString("\n")
	}

	if strings.TrimSpace(result.Stderr) != "" {
		builder.WriteString("Stderr:\n")
		builder.WriteString(result.Stderr)
		builder.WriteString("\n")
	}

	if err != nil {
		builder.WriteString(fmt.Sprintf("Error: %v\n", err))
	}

	builder.WriteString("-------------------------")

	return builder.String()
}

func formatCountResult(plan model.Plan, result model.ExecutionResult) string {
	output := strings.TrimSpace(result.Stdout)
	if output == "" {
		return fmt.Sprintf("No output was produced while counting matches in %s.", plan.Target)
	}

	if _, err := strconv.Atoi(output); err != nil {
		return output
	}

	patternDescription := joinPatterns(plan.Patterns)

	return fmt.Sprintf(
		"The number of %s matches found in %s is %s.",
		patternDescription,
		plan.Target,
		output,
	)
}

func formatListResult(plan model.Plan, result model.ExecutionResult) string {
	output := strings.TrimSpace(result.Stdout)

	if output == "" {
		if plan.Intent == "find_text" {
			return fmt.Sprintf("No matching text was found in %s.", plan.Target)
		}

		patternDescription := joinPatterns(plan.Patterns)
		return fmt.Sprintf("No %s matches were found in %s.", patternDescription, plan.Target)
	}

	if plan.Intent == "find_text" {
		return fmt.Sprintf("Matches found in %s:\n%s", plan.Target, output)
	}

	patternDescription := joinPatterns(plan.Patterns)
	return fmt.Sprintf("Matches found for %s in %s:\n%s", patternDescription, plan.Target, output)
}

func formatSimpleOperationResult(result model.ExecutionResult, fallbackMessage string) string {
	output := strings.TrimSpace(result.Stdout)
	if output != "" {
		return output
	}

	return fallbackMessage
}

func formatDefaultResult(result model.ExecutionResult) string {
	if strings.TrimSpace(result.Stdout) != "" {
		return result.Stdout
	}

	if strings.TrimSpace(result.Stderr) != "" {
		return result.Stderr
	}

	return "The command finished with no output."
}

func joinPatterns(patterns []string) string {
	switch len(patterns) {
	case 0:
		return "log"
	case 1:
		return patterns[0]
	case 2:
		return patterns[0] + " and " + patterns[1]
	default:
		return strings.Join(patterns[:len(patterns)-1], ", ") + ", and " + patterns[len(patterns)-1]
	}
}

func formatCountFilesResult(plan model.Plan, result model.ExecutionResult) string {
	output := strings.TrimSpace(result.Stdout)
	if output == "" {
		return fmt.Sprintf("No output was produced while counting files in %s.", plan.Target)
	}

	if _, err := strconv.Atoi(output); err != nil {
		return output
	}

	return fmt.Sprintf(
		"The number of files found in %s is %s.",
		plan.Target,
		output,
	)
}

func formatListFilesResult(plan model.Plan, result model.ExecutionResult) string {
	output := strings.TrimSpace(result.Stdout)
	if output == "" {
		return fmt.Sprintf("No files were found in %s.", plan.Target)
	}

	return fmt.Sprintf("Files found in %s:\n%s", plan.Target, output)
}
