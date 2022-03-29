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
		if len(m.Parameters) == 0 {
			str += "void."
		} else {
			for i, p := range m.Parameters {
				if i > 0 {
					str += ", "
				}
				str += p.String()
			}
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
	BatchSize       int                         `json:"batchSize"`
	NumOfEpochs     int                         `json:"numOfEpochs"`
	GenerationTasks MethodGenerationTaskOptions `json:"generationTasks"`
}

type MethodGenerationTaskOptions struct {
	// Defines, which tasks should also be performed when generating parameter names in the same task
	ParameterNames CompounTaskOptions `json:"parameterNames"`
	// If true, parameter type generation is performed in a separate task
	ParameterTypes bool `json:"parameterTypes"`
	// If true, return type generation is performed in a separate task
	ReturnType bool `json:"returnType"`
}

type CompounTaskOptions struct {
	// If true, the parameter list generation will be extended by return type generation in the same task
	WithReturnType bool `json:"withReturnType"`
	// If true, the parameter list generation will be extended by parameter type generation in the same task
	WithParameterTypes bool `json:"withParameterTypes"`
}
