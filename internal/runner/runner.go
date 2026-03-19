package runner

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"codeAct-poc/internal/planner"
)

type Result struct {
	Value  string
	Output string
}

func Run(plan planner.Plan) (Result, error) {
	switch plan.Action {
	case "search_logs":
		return runSearchLogs(plan)
	default:
		return Result{}, errors.New("unsupported action")
	}
}

func runSearchLogs(plan planner.Plan) (Result, error) {
	info, err := os.Stat(plan.Target)
	if err != nil {
		if os.IsNotExist(err) {
			return Result{}, errors.New("target path does not exist")
		}
		return Result{}, err
	}

	if !info.IsDir() {
		return Result{}, errors.New("target path is not a directory")
	}

	switch plan.Mode {
	case "count":
		total, err := countMatchesInDirectory(plan.Target, plan.Patterns)
		if err != nil {
			return Result{}, err
		}

		value := strconv.Itoa(total)

		return Result{
			Value:  value,
			Output: value,
		}, nil

	case "list":
		lines, err := listMatchesInDirectory(plan.Target, plan.Patterns)
		if err != nil {
			return Result{}, err
		}

		output := strings.Join(lines, "\n")

		return Result{
			Value:  strconv.Itoa(len(lines)),
			Output: output,
		}, nil

	default:
		return Result{}, errors.New("unsupported mode")
	}
}

func countMatchesInDirectory(target string, patterns []string) (int, error) {
	total := 0

	entries, err := os.ReadDir(target)
	if err != nil {
		return 0, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(target, entry.Name())

		count, err := countMatchesInFile(filePath, patterns)
		if err != nil {
			return 0, err
		}

		total += count
	}

	return total, nil
}

func countMatchesInFile(filePath string, patterns []string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	count := 0
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.ToLower(scanner.Text())

		for _, pattern := range patterns {
			if strings.Contains(line, strings.ToLower(pattern)) {
				count++
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	return count, nil
}

func listMatchesInDirectory(target string, patterns []string) ([]string, error) {
	var results []string

	entries, err := os.ReadDir(target)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(target, entry.Name())

		lines, err := listMatchesInFile(filePath, patterns)
		if err != nil {
			return nil, err
		}

		results = append(results, lines...)
	}

	return results, nil
}

func listMatchesInFile(filePath string, patterns []string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var matches []string
	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		originalLine := scanner.Text()
		lowerLine := strings.ToLower(originalLine)

		for _, pattern := range patterns {
			if strings.Contains(lowerLine, strings.ToLower(pattern)) {
				matches = append(matches, filePath+":"+strconv.Itoa(lineNumber)+": "+originalLine)
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return matches, nil
}
