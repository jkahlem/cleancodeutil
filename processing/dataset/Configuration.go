package dataset

type Configuration struct {
	Name       string                  `json:"name"`
	Model      ModelConfiguration      `json:"model"`
	Filters    []string                `json:"filters"`
	Evaluation EvaluationConfiguration `json:"evaluation"`
	Subsets    []Configuration         `json:"subsets"`
}

type EvaluationConfiguration struct {
}

type ModelConfiguration struct {
	Type ModelType `json:"type"`
}
