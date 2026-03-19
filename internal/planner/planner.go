package planner

import (
	"runtime"

	"codeAct-poc/internal/ai"
	"codeAct-poc/internal/fallback"
	"codeAct-poc/internal/model"
)

func Build(task string) (model.Plan, error) {
	language := detectScriptLanguage()

	plan, err := ai.GeneratePlan(task, language)
	if err == nil {
		plan.Source = "ai"
		return plan, nil
	}

	plan, err = fallback.GeneratePlan(task, language)
	if err != nil {
		return model.Plan{}, err
	}

	plan.Source = "fallback"
	return plan, nil
}

func detectScriptLanguage() string {
	if runtime.GOOS == "windows" {
		return "powershell"
	}

	return "bash"
}
