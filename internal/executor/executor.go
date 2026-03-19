package executor

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	"codeAct-poc/internal/model"
)

func Execute(plan model.Plan) (model.ExecutionResult, error) {
	scriptPath, err := writeTempScript(plan.Language, plan.Script)
	if err != nil {
		return model.ExecutionResult{}, err
	}

	defer os.Remove(scriptPath)

	command, args, err := buildCommand(plan.Language, scriptPath)
	if err != nil {
		return model.ExecutionResult{}, err
	}

	cmd := exec.Command(command, args...)

	outputBytes, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(outputBytes))

	result := model.ExecutionResult{
		Stdout:   "",
		Stderr:   "",
		ExitCode: extractExitCode(err),
	}

	if err == nil {
		result.Stdout = output
		return result, nil
	}

	result.Stderr = output
	return result, err
}

func writeTempScript(language string, script string) (string, error) {
	pattern := "agent-script-*"

	switch language {
	case "powershell":
		pattern += ".ps1"
	case "bash":
		pattern += ".sh"
	default:
		return "", fmt.Errorf("unsupported script language: %s", language)
	}

	file, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if language == "bash" {
		if _, err := file.WriteString("#!/usr/bin/env bash\n"); err != nil {
			return "", err
		}
	}

	if _, err := file.WriteString(script); err != nil {
		return "", err
	}

	if language == "bash" {
		if err := os.Chmod(file.Name(), 0o700); err != nil {
			return "", err
		}
	}

	return file.Name(), nil
}

func buildCommand(language string, scriptPath string) (string, []string, error) {
	switch language {
	case "powershell":
		if runtime.GOOS == "windows" {
			return "powershell", []string{
				"-NoProfile",
				"-ExecutionPolicy", "Bypass",
				"-File", scriptPath,
			}, nil
		}

		return "pwsh", []string{
			"-NoProfile",
			"-File", scriptPath,
		}, nil

	case "bash":
		return "bash", []string{scriptPath}, nil

	default:
		return "", nil, fmt.Errorf("unsupported script language: %s", language)
	}
}

func extractExitCode(err error) int {
	if err == nil {
		return 0
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}

	return 1
}
