package formatter

import (
	"fmt"
	"strings"

	"codeAct-poc/internal/model"
)

func Format(plan model.Plan, result model.ExecutionResult) string {
	switch plan.Intent {
	case "count_matches":
		return formatCountResult(plan, result)
	case "list_matches":
		return formatListResult(plan, result)
	default:
		if strings.TrimSpace(result.Stdout) != "" {
			return result.Stdout
		}

		if strings.TrimSpace(result.Stderr) != "" {
			return result.Stderr
		}

		return "The command finished with no output."
	}
}

func formatCountResult(plan model.Plan, result model.ExecutionResult) string {
	if strings.TrimSpace(result.Stdout) == "" {
		return fmt.Sprintf(
			"No output was produced while counting matches in %s.",
			plan.Target,
		)
	}

	patternDescription := joinPatterns(plan.Patterns)

	return fmt.Sprintf(
		"The number of %s matches found in %s is %s.",
		patternDescription,
		plan.Target,
		result.Stdout,
	)
}

func formatListResult(plan model.Plan, result model.ExecutionResult) string {
	patternDescription := joinPatterns(plan.Patterns)
	output := strings.TrimSpace(result.Stdout)

	if output == "" {
		return fmt.Sprintf(
			"No %s matches were found in %s.",
			patternDescription,
			plan.Target,
		)
	}

	return fmt.Sprintf(
		"Matches found for %s in %s:\n%s",
		patternDescription,
		plan.Target,
		output,
	)
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

func FormatDebugPlan(plan model.Plan) string {
	return fmt.Sprintf(
		"---- PLAN DEBUG ----\nUseCase: %s\nIntent: %s\nLanguage: %s\nTarget: %s\nPatterns: %v\nSource: %s\nScript:\n%s\n--------------------",
		plan.UseCase,
		plan.Intent,
		plan.Language,
		plan.Target,
		plan.Patterns,
		plan.Source,
		plan.Script,
	)
}

func FormatDebugResult(result model.ExecutionResult, err error) string {
	builder := strings.Builder{}

	builder.WriteString("---- EXECUTION DEBUG ----\n")

	builder.WriteString(fmt.Sprintf("ExitCode: %d\n", result.ExitCode))

	if result.Stdout != "" {
		builder.WriteString("Stdout:\n")
		builder.WriteString(result.Stdout + "\n")
	}

	if result.Stderr != "" {
		builder.WriteString("Stderr:\n")
		builder.WriteString(result.Stderr + "\n")
	}

	if err != nil {
		builder.WriteString(fmt.Sprintf("Error: %v\n", err))
	}

	builder.WriteString("-------------------------")

	return builder.String()
}
