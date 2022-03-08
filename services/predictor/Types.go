package predictor

type Evaluation struct {
	AccScore float64 `json:"accScore"`
	EvalLoss float64 `json:"evalLoss"`
	F1Score  float64 `json:"f1Score"`
	MCC      float64 `json:"mcc"`
}

type MethodContext struct {
	MethodName string
	ClassName  string
	IsStatic   bool
}

type MethodValues struct {
	ReturnType string
	Parameters string
}
