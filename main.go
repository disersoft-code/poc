package main

import (
	"codeAct-poc/internal/cli"
	//"codeAct-poc/internal/executor"
	"codeAct-poc/internal/formatter"
	//"codeAct-poc/internal/planner"
	"codeAct-poc/internal/agent"

	//"codeAct-poc/internal/runner"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("No .env file found")
	}

	input, err := cli.ParseInput()
	if err != nil {
		fmt.Println("Error on parsing input:", err)
		os.Exit(1)
	}

	if input.DryRun {
		plan, err := agent.Build(input.Task)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		if input.Debug {
			fmt.Println(formatter.FormatDebugPlan(plan))
		}

		fmt.Println("Generated script:")
		fmt.Println(plan.Script)
		return
	}

	runResult, err := agent.Run(input.Task)

	if input.Debug {
		fmt.Println(formatter.FormatDebugPlan(runResult.Plan))
		fmt.Println(formatter.FormatDebugResult(runResult.Result, err))
	}

	if err != nil {
		if runResult.Result.Stderr != "" {
			fmt.Println("Error:", runResult.Result.Stderr)
		} else {
			fmt.Println("Error:", err)
		}
		os.Exit(1)
	}

	message := formatter.Format(runResult.Plan, runResult.Result)
	fmt.Println(message)

	/*
		plan, err := planner.Build(input.Task)
		if err != nil {
			fmt.Println("Error on building plan:", err)
			os.Exit(1)
		}

		if input.Debug {
			fmt.Println(formatter.FormatDebugPlan(plan))
		}

		if input.DryRun {
			fmt.Println("Generated script:")
			fmt.Println(plan.Script)
			return
		}

		result, err := executor.Execute(plan)

		if input.Debug {
			fmt.Println(formatter.FormatDebugResult(result, err))
		}

		if err != nil {
			if result.Stderr != "" {
				fmt.Println("Error:", result.Stderr)
			} else {
				fmt.Println("Error:", err)
			}
			os.Exit(1)
		}

		message := formatter.Format(plan, result)
		fmt.Println(message)
	*/
}
