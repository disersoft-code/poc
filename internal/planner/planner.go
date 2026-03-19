package planner

import (
	"errors"
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

	plan, fallbackErr := fallback.GeneratePlan(task, language)
	if fallbackErr == nil {
		plan.Source = "fallback"
		return plan, nil
	}

	return model.Plan{}, errors.New("this command is not supported without AI")
}

func detectScriptLanguage() string {
	if runtime.GOOS == "windows" {
		return "powershell"
	}

	return "bash"
}
