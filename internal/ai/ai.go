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

	if err := validatePlan(plan, language); err != nil {
		return model.Plan{}, err
	}

	return plan, nil
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
		return model.Plan{}, fmt.Errorf("model repair request failed with status %d", resp.StatusCode)
	}

	var responseBody chatCompletionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return model.Plan{}, err
	}

	if len(responseBody.Choices) == 0 {
		return model.Plan{}, errors.New("model returned no choices for repair")
	}

	content := strings.TrimSpace(responseBody.Choices[0].Message.Content)
	if content == "" {
		return model.Plan{}, errors.New("model returned empty repair content")
	}

	content = stripCodeFence(content)

	var plan model.Plan
	if err := json.Unmarshal([]byte(content), &plan); err != nil {
		return model.Plan{}, fmt.Errorf("invalid repaired plan json: %w", err)
	}

	if err := validatePlan(plan, previousPlan.Language); err != nil {
		return model.Plan{}, err
	}

	return plan, nil
}

func buildSystemPrompt(language string) string {
	return fmt.Sprintf(`You generate executable scripts for a Go CodeAct-style log agent.

Return only valid JSON.
Do not wrap the JSON in markdown.
Do not include explanations.

The JSON schema is:
{
  "use_case": "log_agent",
  "intent": "count_matches" or "list_matches",
  "language": "%s",
  "target": "path string",
  "patterns": ["error", "warning"],
  "script": "full executable script"
}

Rules:
- The use case must always be "log_agent".
- The language must always be "%s".
- Supported intents: "count_matches", "list_matches".
- The target must come from the user task.
- Patterns should include "error" and/or "warning". If unclear, default to ["error"].
- The script must be executable in the requested language.
- The script must only search files in the target directory, not delete, rename, or modify files.
- For count_matches, the script must output only the final count.
- For list_matches, the script must output one match per line in the format file:line: content.
`, language, language)
}

func buildRepairSystemPrompt(language string) string {
	return fmt.Sprintf(`You repair executable scripts for a Go CodeAct-style log agent.

Return only valid JSON.
Do not wrap the JSON in markdown.
Do not include explanations.

The JSON schema is:
{
  "use_case": "log_agent",
  "intent": "count_matches" or "list_matches",
  "language": "%s",
  "target": "path string",
  "patterns": ["error", "warning"],
  "script": "full executable script"
}

Rules:
- The use case must always be "log_agent".
- The language must always be "%s".
- Supported intents: "count_matches", "list_matches".
- The script must only search files in the target directory.
- Do not delete, rename, or modify files.
- Fix the script using the execution error feedback.
- For count_matches, the script must output only the final count.
- For list_matches, the script must output one match per line in the format file:line: content.
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
Patterns: %v

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
		previousPlan.Patterns,
		previousPlan.Script,
		executionResult.Stdout,
		executionResult.Stderr,
		executionResult.ExitCode,
	)
}

func buildUserPrompt(task string, language string) string {
	return fmt.Sprintf(`Task: %s

Generate a JSON plan for the requested task using %s.`, task, language)
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
	if plan.UseCase != "log_agent" {
		return errors.New("invalid plan use_case")
	}

	if plan.Intent != "count_matches" && plan.Intent != "list_matches" {
		return errors.New("invalid plan intent")
	}

	if plan.Language != expectedLanguage {
		return errors.New("invalid plan language")
	}

	if strings.TrimSpace(plan.Target) == "" {
		return errors.New("invalid plan target")
	}

	if len(plan.Patterns) == 0 {
		return errors.New("invalid plan patterns")
	}

	if strings.TrimSpace(plan.Script) == "" {
		return errors.New("invalid plan script")
	}

	return nil
}
