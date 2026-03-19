# CodeAct-Style Agent in Go (PoC)

## Overview

This project implements a **CodeAct-style agent** in Go that executes tasks by generating and running scripts (PowerShell or Bash) instead of relying on predefined function calls.

The agent interprets a natural language task, generates an executable script using an LLM (or fallback logic), runs it, and uses execution feedback to improve the result.

This is a **proof of concept**, not a production-ready system.

---

## Key Concepts

### Code-as-Action

Instead of calling predefined APIs, the agent:

1. Generates executable code (script)
2. Runs it in the environment
3. Uses the output to validate or improve the result

This enables:

- multi-step logic
- conditionals
- real interaction with the system (filesystem, CLI)

---

## Features

- CLI-based interface
- Supports:
  - Log analysis (`log_agent`)
  - File system operations (`file_agent`)
- Dynamic script generation using AI
- Automatic OS detection:
  - Windows → PowerShell
  - Linux/macOS → Bash
- Retry mechanism using execution feedback
- Fallback mode when AI is unavailable
- Debug mode (`--debug`)
- Dry-run mode (`--dry-run`)

---

## Architecture

```text
CLI → Planner → (AI or Fallback) → Plan → Executor → Formatter → Output
                         ↓
                     Retry (AI repair)
```

### Components

- **cli** → Parses user input and flags
- **planner** → Decides whether to use AI or fallback
- **ai** → Generates plans (scripts) using an LLM
- **fallback** → Handles basic log queries without AI
- **executor** → Runs scripts and captures output
- **formatter** → Produces user-friendly output
- **agent** → Orchestrates execution, retry, and fallback

---

## Installation

```bash
go mod tidy
```

---

## Configuration

Create a `.env` file:

```env
AGENT_MODEL_API_URL=https://api.openai.com/v1/chat/completions
AGENT_MODEL_NAME=gpt-4o-mini
AGENT_MODEL_API_KEY=your_api_key
```

---

## Usage

### Basic command

```bash
go run . --task "count errors in ./logs"
```

### Debug mode

```bash
go run . --task "count errors in ./logs" --debug
```

Shows:

- generated plan
- script
- execution result

### Dry run (no execution)

```bash
go run . --task "create a file named notes.txt in ./output with content hello world" --dry-run
```

---

## Examples

### Log analysis

```bash
go run . --task "count errors and warnings in ./logs"
```

```bash
go run . --task "list errors in ./logs"
```

### File operations (AI required)

```bash
go run . --task "create a file named notes.txt in ./output with content hello world"
```

```bash
go run . --task "delete ./output/notes.txt"
```

```bash
go run . --task "rename ./a.txt to ./b.txt"
```

```bash
go run . --task "replace foo with bar in ./file.txt"
```

---

## Behavior Without AI

If AI is not configured:

- Supported:
  - count/list errors and warnings in logs
- Not supported:
  - file system operations

Example:

```text
This command is not supported without AI.
```

---

## Execution Model

1. Build plan (AI or fallback)
2. Execute script
3. If execution fails:
   - Retry using AI with error feedback
4. If still failing:
   - Fallback (if applicable)

---

## Output Rules

- Scripts must print only the final user-facing result
- No intermediate command output
- Validation outcomes are treated as normal results:
  - "File does not exist."
  - "No matches found."
- Only real execution failures trigger errors:
  - syntax errors
  - invalid commands
  - permission issues

---

## Limitations

- No sandboxing (scripts run locally)
- Limited validation of AI-generated scripts
- Error detection relies partially on heuristics
- No concurrency or parallel execution
- Limited natural language understanding (prompt-based)

---

## Future Improvements

- Safer execution (sandboxing)
- Better stdout/stderr separation
- Structured outputs instead of plain text
- Multi-step planning
- Support for additional tools (APIs, databases)
- More robust validation of generated scripts

---

## Summary

This PoC demonstrates:

- Code-as-action paradigm
- Dynamic script generation
- Execution-feedback loop
- Hybrid AI + deterministic fallback design

It shows how an agent can interact with real environments using executable code instead of predefined APIs.