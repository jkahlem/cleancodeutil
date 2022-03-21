package predictor

import "fmt"

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

func (m MethodContext) String() string {
	str := ""
	if m.IsStatic {
		str += "static "
	}
	if m.ClassName != "" {
		str += m.ClassName + "."
	}
	return str + string(m.MethodName)
}

type MethodValues struct {
	ReturnType string      `json:"returnType"`
	Parameters []Parameter `json:"parameters"`
}

func (m MethodValues) String() string {
	str := ""
	if len(m.Parameters) > 0 {
		str += "parameters: "
		for i, p := range m.Parameters {
			if i > 0 {
				str += ", "
			}
			str += p.String()
		}
	}
	if m.ReturnType != "" {
		if str != "" {
			str += ". "
		}
		str += "returns: " + m.ReturnType
	}
	return str
}

type Parameter struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (p Parameter) String() string {
	return fmt.Sprintf("%s %s", p.Type, p.Name)
}

type Method struct {
	Context MethodContext `json:"context"`
	Values  MethodValues  `json:"values"`
}

type Options struct {
	Identifier   string          `json:"identifier"`
	LabelsCsv    string          `json:"labels"`
	Type         SupportedModels `json:"type"`
	ModelOptions ModelOptions    `json:"modelOptions"`
}

type ModelOptions struct {
	BatchSize   int `json:"batchSize"`
	NumOfEpochs int `json:"numOfEpochs"`
}
