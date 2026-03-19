package model

type Plan struct {
	UseCase  string   `json:"use_case"`
	Intent   string   `json:"intent"`
	Language string   `json:"language"`
	Target   string   `json:"target"`
	Patterns []string `json:"patterns"`
	Script   string   `json:"script"`
	Source   string   `json:"source"`
}
