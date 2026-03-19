package agent

import (
	"fmt"

	"codeAct-poc/internal/ai"
	"codeAct-poc/internal/executor"
	"codeAct-poc/internal/fallback"
	"codeAct-poc/internal/model"
	"codeAct-poc/internal/planner"
)

type RunResult struct {
	Plan   model.Plan
	Result model.ExecutionResult
}

func Build(task string) (model.Plan, error) {
	return planner.Build(task)
}

func Run(task string) (RunResult, error) {
	plan, err := Build(task)
	if err != nil {
		return RunResult{}, err
	}

	result, err := executor.Execute(plan)
	if err == nil {
		return RunResult{
			Plan:   plan,
			Result: result,
		}, nil
	}

	if plan.Source == "ai" {
		correctedPlan, retryErr := ai.RegeneratePlan(task, plan, result)
		if retryErr == nil {
			correctedPlan.Source = "ai_retry"

			retryResult, retryExecErr := executor.Execute(correctedPlan)
			if retryExecErr == nil {
				return RunResult{
					Plan:   correctedPlan,
					Result: retryResult,
				}, nil
			}
		}
	}

	fallbackPlan, fallbackErr := fallback.GeneratePlan(task, plan.Language)
	if fallbackErr != nil {
		return RunResult{
			Plan:   plan,
			Result: result,
		}, fmt.Errorf("execution failed and fallback plan could not be generated: %w", err)
	}
	fallbackPlan.Source = "fallback"

	fallbackResult, fallbackExecErr := executor.Execute(fallbackPlan)
	if fallbackExecErr != nil {
		return RunResult{
			Plan:   fallbackPlan,
			Result: fallbackResult,
		}, fallbackExecErr
	}

	return RunResult{
		Plan:   fallbackPlan,
		Result: fallbackResult,
	}, nil
}
