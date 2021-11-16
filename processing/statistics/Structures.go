package statistics

import (
	"returntypes-langserver/common/csv"
	"returntypes-langserver/services/predictor"
)

type Statistics struct {
	Projects   []ProjectStatistics   `json:"projects"`
	General    GeneralStatistics     `json:"general"`
	Dataset    SpecificStatistics    `json:"dataset"`
	TestCode   SpecificStatistics    `json:"testCode"`
	Labels     LabelStatistics       `json:"labels"`
	Evaluation *predictor.Evaluation `json:"evaluation,omitempty"`
}

type LabelStatistics struct {
	Getter      int `json:"getter"`
	Setter      int `json:"setter"`
	ChainMethod int `json:"chainMethod"`
	TestCode    int `json:"testCode"`
	Override    int `json:"override"`
	ArrayType   int `json:"arrayType"`
}

type SpecificStatistics struct {
	MethodsCount int                    `json:"methodsCount"`
	ReturnTypes  ReturnTypeUsageCounter `json:"returnTypes"`
}

type GeneralStatistics struct {
	FileCount                     int                           `json:"fileCount"`
	ClassCount                    int                           `json:"classCount"`
	MethodsCount                  int                           `json:"methodsCount"`
	ReturnTypes                   ReturnTypeUsageCounter        `json:"returnTypes"`
	MethodListBeforeSummarization []csv.MethodSummarizationData `json:"methodListBeforeSummarization"`
}

type ProjectStatistics struct {
	FileCount             int                    `json:"fileCount"`
	MethodsCount          int                    `json:"methodsCount"`
	MethodsInDatasetCount int                    `json:"methodsInDatasetCount"`
	Name                  string                 `json:"name"`
	DirName               string                 `json:"dirName"`
	Description           string                 `json:"description"`
	ReturnTypes           ReturnTypeUsageCounter `json:"returnTypes"`
}
