package predictor

type Evaluation struct {
	AccScore float64 `json:"accScore" mapstructure:"accScore"`
	EvalLoss float64 `json:"evalLoss" mapstructure:"evalLoss"`
	F1Score  float64 `json:"f1Score" mapstructure:"f1Score"`
	MCC      float64 `json:"mcc" mapstructure:"mcc"`
}

type MethodTypeMap map[PredictableMethodName]string
