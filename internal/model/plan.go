package model

type Plan struct {
	UseCase     string   `json:"use_case"`
	Intent      string   `json:"intent"`
	Language    string   `json:"language"`
	Target      string   `json:"target"`
	Destination string   `json:"destination"`
	Patterns    []string `json:"patterns"`
	Replacement string   `json:"replacement"`
	Content     string   `json:"content"`
	Script      string   `json:"script"`
	Source      string   `json:"source"`
}
