package predictor

type Evaluation struct {
	AccScore float64 `json:"accScore"`
	EvalLoss float64 `json:"evalLoss"`
	F1Score  float64 `json:"f1Score"`
	MCC      float64 `json:"mcc"`
}

type MethodContext struct {
	MethodName PredictableMethodName `json:"methodName"`
	ClassName  string                `json:"className"`
	IsStatic   bool                  `json:"isStatic"`
	Types      []string              `json:"types"`
}

type MethodValues struct {
	ReturnType string      `json:"returnType"`
	Parameters []Parameter `json:"parameters"`
}

type Parameter struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Method struct {
	Context MethodContext `json:"context"`
	Values  MethodValues  `json:"values"`
}

type Options struct {
	Identifier   string          `json:"identifier"`
	LabelsCsv    string          `json:"labels"`
	Type         SupportedModels `json:"type"`
	ModelOptions ModelOptions    `json:"model"`
}

type ModelOptions struct {
}
