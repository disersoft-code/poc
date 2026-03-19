package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"codeAct-poc/internal/model"
)

type chatCompletionsRequest struct {
	Model       string                   `json:"model"`
	Messages    []chatCompletionsMessage `json:"messages"`
	Temperature float64                  `json:"temperature"`
}

type chatCompletionsMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionsResponse struct {
	Choices []struct {
		Message chatCompletionsMessage `json:"message"`
	} `json:"choices"`
}

func GeneratePlan(task string, language string) (model.Plan, error) {
	apiURL := strings.TrimSpace(os.Getenv("AGENT_MODEL_API_URL"))
	apiKey := strings.TrimSpace(os.Getenv("AGENT_MODEL_API_KEY"))
	modelName := strings.TrimSpace(os.Getenv("AGENT_MODEL_NAME"))

	if apiURL == "" {
		return model.Plan{}, errors.New("missing AGENT_MODEL_API_URL")
	}

	if modelName == "" {
		return model.Plan{}, errors.New("missing AGENT_MODEL_NAME")
	}

	systemPrompt := buildSystemPrompt(language)
	userPrompt := buildUserPrompt(task, language)

	requestBody := chatCompletionsRequest{
		Model: modelName,
		Messages: []chatCompletionsMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
		Temperature: 0.1,
	}

	return sendPlanRequest(apiURL, apiKey, requestBody, language)
}

func RegeneratePlan(task string, previousPlan model.Plan, executionResult model.ExecutionResult) (model.Plan, error) {
	apiURL := strings.TrimSpace(os.Getenv("AGENT_MODEL_API_URL"))
	apiKey := strings.TrimSpace(os.Getenv("AGENT_MODEL_API_KEY"))
	modelName := strings.TrimSpace(os.Getenv("AGENT_MODEL_NAME"))

	if apiURL == "" {
		return model.Plan{}, errors.New("missing AGENT_MODEL_API_URL")
	}

	if modelName == "" {
		return model.Plan{}, errors.New("missing AGENT_MODEL_NAME")
	}

	systemPrompt := buildRepairSystemPrompt(previousPlan.Language)
	userPrompt := buildRepairUserPrompt(task, previousPlan, executionResult)

	requestBody := chatCompletionsRequest{
		Model: modelName,
		Messages: []chatCompletionsMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
		Temperature: 0.1,
	}

	return sendPlanRequest(apiURL, apiKey, requestBody, previousPlan.Language)
}

func sendPlanRequest(apiURL string, apiKey string, requestBody chatCompletionsRequest, expectedLanguage string) (model.Plan, error) {
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return model.Plan{}, err
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return model.Plan{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return model.Plan{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return model.Plan{}, fmt.Errorf("model request failed with status %d", resp.StatusCode)
	}

	var responseBody chatCompletionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return model.Plan{}, err
	}

	if len(responseBody.Choices) == 0 {
		return model.Plan{}, errors.New("model returned no choices")
	}

	content := strings.TrimSpace(responseBody.Choices[0].Message.Content)
	if content == "" {
		return model.Plan{}, errors.New("model returned empty content")
	}

	content = stripCodeFence(content)

	var plan model.Plan
	if err := json.Unmarshal([]byte(content), &plan); err != nil {
		return model.Plan{}, fmt.Errorf("invalid model plan json: %w", err)
	}

	if err := validatePlan(plan, expectedLanguage); err != nil {
		return model.Plan{}, err
	}

	return plan, nil
}

func buildSystemPrompt(language string) string {
	return fmt.Sprintf(`You generate executable scripts for a Go CodeAct-style agent.

Return only valid JSON.
Do not wrap the JSON in markdown.
Do not include explanations.

The JSON schema is:
{
  "use_case": "log_agent" or "file_agent",
  "intent": "count_matches" or "list_matches" or "count_files" or "list_files" or "create_file" or "delete_path" or "rename_path" or "copy_path" or "replace_text" or "find_text",
  "language": "%s",
  "target": "path string",
  "destination": "path string",
  "patterns": ["pattern1", "pattern2"],
  "replacement": "replacement text",
  "content": "file content",
  "script": "full executable script"
}

Rules:
- The language must always be "%s".
- Supported use cases: "log_agent", "file_agent".
- Supported intents:
  - "count_matches"
  - "list_matches"
  - "count_files"
  - "list_files"
  - "create_file"
  - "delete_path"
  - "rename_path"
  - "copy_path"
  - "replace_text"
  - "find_text"- The target must come from the user task when applicable.
- The destination must be included for copy and rename operations.
- The replacement must be included for replace_text.
- The content must be included for create_file when the user provides content.
- The script must be executable in the requested language.
- The script may interact with the file system.
- The script must not ask for user input.
- Prefer deterministic scripts.
- If the user asks to search logs for errors or warnings, use "log_agent".
- For generic file system actions, use "file_agent".
- PowerShell scripts must set $ErrorActionPreference = 'Stop' at the beginning.
- Bash scripts must use set -e at the beginning.
- Prefer explicit path validation before performing file system operations.
- Prefer Set-Content for file creation in PowerShell.
- For create_file, ensure the parent directory exists before writing the file.
- For rename_path and copy_path, ensure the parent directory of the destination exists when needed.
- Basic validation outcomes must be treated as normal results, not execution errors.
- Missing files or directories should return a short informative message and exit successfully.
- If there is nothing to replace, return a short informative message and exit successfully.
- If no matches are found, return a short informative message and exit successfully.
- Only use a failing exit code for real execution errors such as syntax errors, permission issues, invalid commands, or unexpected runtime failures.
- Do not print both an informative validation message and a success message for the same operation.
- Scripts must print only the final user-facing result unless the task explicitly asks for search results.
- Do not print intermediate command output, tables, objects, or diagnostic information unless the task explicitly asks for it.
- PowerShell scripts must suppress unintended command output, for example by piping New-Item to Out-Null when its output is not part of the final result.
- Bash scripts must suppress unintended command output and print only the requested final result.
- For count_matches, the script must output only the final count.
- For list_matches and find_text, the script must output one match per line in the format file:line: content.
- For create_file, delete_path, rename_path, copy_path, and replace_text, print exactly one short final line describing the outcome.
- For count_matches, output only a numeric count on success.
- If counting cannot be performed because of a validation outcome, output a short informative message instead of a number.
- Use "count_files" when the user asks to count files in a directory without searching file contents.
- For count_files, the script must output only the final numeric file count on success.
- If the directory does not exist for count_files or list_files, return a short informative message and exit successfully.
- Use "list_files" when the user asks to list files in a directory without searching file contents.
- For list_files, output one file path per line.
- For rename_path in PowerShell, prefer using Move-Item with -Destination instead of Rename-Item.
`, language, language)
}

func buildUserPrompt(task string, language string) string {
	return fmt.Sprintf(`Task: %s

Generate a JSON plan for the requested task using %s.`, task, language)
}

func buildRepairSystemPrompt(language string) string {
	return fmt.Sprintf(`You repair executable scripts for a Go CodeAct-style agent.

Return only valid JSON.
Do not wrap the JSON in markdown.
Do not include explanations.

The JSON schema is:
{
  "use_case": "log_agent" or "file_agent",
  "intent": "count_matches" or "list_matches" or "count_files" or "list_files" or "create_file" or "delete_path" or "rename_path" or "copy_path" or "replace_text" or "find_text",
  "language": "%s",
  "target": "path string",
  "destination": "path string",
  "patterns": ["pattern1", "pattern2"],
  "replacement": "replacement text",
  "content": "file content",
  "script": "full executable script"
}

Rules:
- The language must always be "%s".
- Fix the script using the execution feedback.
- Keep the same overall task intent.
- Return a corrected executable script.
- PowerShell scripts must set $ErrorActionPreference = 'Stop' at the beginning.
- Bash scripts must use set -e at the beginning.
- Ensure parent directories exist when needed.
- Basic validation outcomes must be treated as normal results, not execution errors.
- Missing files or directories should return a short informative message and exit successfully.
- If there is nothing to replace, return a short informative message and exit successfully.
- If no matches are found, return a short informative message and exit successfully.
- Only use a failing exit code for real execution errors such as syntax errors, permission issues, invalid commands, or unexpected runtime failures.
- Avoid turning expected validation outcomes into hard failures.
- Use the execution feedback to avoid repeating the same mistake.
- The corrected script must print only the final user-facing result unless the task explicitly asks for search results.
- Suppress unintended command output.
- Do not print intermediate command output, tables, objects, or diagnostic information unless the task explicitly asks for it.
- For create_file, delete_path, rename_path, copy_path, and replace_text, print exactly one short final line describing the outcome.
- Do not print both an informative validation message and a success message for the same operation.
- Keep "count_files" for directory file counting tasks that do not require content search.
- Keep "list_files" for directory file listing tasks that do not require content search.
- For list_files, the corrected script must output one file path per line.
`, language, language)
}

func buildRepairUserPrompt(task string, previousPlan model.Plan, executionResult model.ExecutionResult) string {
	return fmt.Sprintf(`The previous generated plan failed during execution.

Original task:
%s

Previous plan:
Use case: %s
Intent: %s
Language: %s
Target: %s
Destination: %s
Patterns: %v
Replacement: %s
Content: %s

Previous script:
%s

Execution stdout:
%s

Execution stderr:
%s

Exit code:
%d

Generate a corrected JSON plan.`, task,
		previousPlan.UseCase,
		previousPlan.Intent,
		previousPlan.Language,
		previousPlan.Target,
		previousPlan.Destination,
		previousPlan.Patterns,
		previousPlan.Replacement,
		previousPlan.Content,
		previousPlan.Script,
		executionResult.Stdout,
		executionResult.Stderr,
		executionResult.ExitCode,
	)
}

func stripCodeFence(value string) string {
	value = strings.TrimSpace(value)

	if strings.HasPrefix(value, "```") {
		lines := strings.Split(value, "\n")
		if len(lines) >= 2 {
			lines = lines[1:]
		}
		if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "```" {
			lines = lines[:len(lines)-1]
		}
		return strings.TrimSpace(strings.Join(lines, "\n"))
	}

	return value
}

func validatePlan(plan model.Plan, expectedLanguage string) error {
	if plan.UseCase != "log_agent" && plan.UseCase != "file_agent" {
		return errors.New("invalid plan use_case")
	}

	switch plan.Intent {
	case "count_matches", "list_matches", "count_files", "list_files", "create_file", "delete_path", "rename_path", "copy_path", "replace_text", "find_text":
	default:
		return errors.New("invalid plan intent")
	}

	if plan.Language != expectedLanguage {
		return errors.New("invalid plan language")
	}

	if strings.TrimSpace(plan.Script) == "" {
		return errors.New("invalid plan script")
	}

	switch plan.Intent {
	case "count_matches", "list_matches", "find_text":
		if strings.TrimSpace(plan.Target) == "" {
			return errors.New("invalid plan target")
		}
		if len(plan.Patterns) == 0 {
			return errors.New("invalid plan patterns")
		}

	case "count_files", "list_files":
		if strings.TrimSpace(plan.Target) == "" {
			return errors.New("invalid plan target")
		}

	case "create_file":
		if strings.TrimSpace(plan.Target) == "" {
			return errors.New("invalid plan target")
		}

	case "delete_path":
		if strings.TrimSpace(plan.Target) == "" {
			return errors.New("invalid plan target")
		}

	case "rename_path", "copy_path":
		if strings.TrimSpace(plan.Target) == "" {
			return errors.New("invalid plan target")
		}
		if strings.TrimSpace(plan.Destination) == "" {
			return errors.New("invalid plan destination")
		}

	case "replace_text":
		if strings.TrimSpace(plan.Target) == "" {
			return errors.New("invalid plan target")
		}
		if len(plan.Patterns) == 0 {
			return errors.New("invalid plan patterns")
		}
		if strings.TrimSpace(plan.Replacement) == "" {
			return errors.New("invalid plan replacement")
		}
	}

	return nil
}
