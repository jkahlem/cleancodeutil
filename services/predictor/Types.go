package predictor

import (
	"fmt"
	"strings"
)

type Evaluation struct {
	AccScore float64 `json:"accScore"`
	EvalLoss float64 `json:"evalLoss"`
	F1Score  float64 `json:"f1Score"`
	MCC      float64 `json:"mcc"`
}

type MethodContext struct {
	MethodName string   `json:"methodName"`
	ClassName  []string `json:"className"`
	IsStatic   bool     `json:"isStatic"`
	Types      []string `json:"types"`
}

func (m MethodContext) String() string {
	str := ""
	if m.IsStatic {
		str += "static "
	}
	if len(m.ClassName) != 0 {
		str += strings.Join(m.ClassName, ".") + "."
	}
	return str + string(m.MethodName)
}

type MethodValues struct {
	ReturnType string      `json:"returnType"`
	Parameters []Parameter `json:"parameters"`
}

type Parameter struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	IsArray bool   `json:"isArray"`
}

func (p Parameter) String() string {
	if p.IsArray {
		return fmt.Sprintf("%s [arr] [tsp] %s", p.Type, p.Name)
	}
	return fmt.Sprintf("%s [tsp] %s", p.Type, p.Name)
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
	Checkpoint   string          `json:"checkpoint"`
}

type ModelOptions struct {
	BatchSize                   int                   `json:"batchSize"`
	NumOfEpochs                 int                   `json:"numOfEpochs"`
	NumReturnSequences          int                   `json:"numReturnSequences"`
	MaxSequenceLength           int                   `json:"maxSequenceLength"`
	DefaultContextTypes         []string              `json:"defaultContext,omitempty"`
	EmptyParameterListByKeyword bool                  `json:"emptyParameterListByKeyword"`
	Adafactor                   Adafactor             `json:"adafactor"`
	Adam                        Adam                  `json:"adam"`
	ModelType                   string                `json:"modelType"`
	ModelName                   string                `json:"modelName"`
	NumBeams                    int                   `json:"numBeams"`
	LengthPenalty               *float64              `json:"lengthPenalty,omitempty"`
	TopK                        *float64              `json:"topK,omitempty"`
	TopN                        *float64              `json:"topN,omitempty"`
	OutputOrder                 *OutputComponentOrder `json:"outputOrder,omitempty"`
}

type OutputComponentOrder struct {
	ReturnType    int `json:"returnType"`
	ParameterName int `json:"parameterName"`
	ParameterType int `json:"parameterType"`
}

type Adafactor struct {
	Beta           *float64  `json:"beta,omitempty"`
	ClipThreshold  *float64  `json:"clipThreshold,omitempty"`
	DecayRate      *float64  `json:"decayRate,omitempty"`
	Eps            []float64 `json:"eps,omitempty"`
	RelativeStep   *bool     `json:"relativeStep,omitempty"`
	WarmupInit     *bool     `json:"warmupInit,omitempty"`
	ScaleParameter *bool     `json:"scaleParameter,omitempty"`
}

type Adam struct {
	LearningRate *float64 `json:"learningRate,omitempty"`
	Eps          *float64 `json:"eps,omitempty"`
}

type Model struct {
	ModelName   string
	Checkpoints []string
}
