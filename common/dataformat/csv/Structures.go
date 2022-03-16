package csv

import "time"

// Generate Marshal / Unmarshal methods (-> Marshaller.go)
//go:generate go run ./marshallerGenerator

type Method struct {
	ClassName  string `excel:"Class name,width=25"`
	ReturnType string `excel:"Return type,width=20"`
	MethodName string `excel:"Method name,width=30"`
	// Parameters are in this format: "<type> <name>" (seperated by a single space)
	Parameters []string `excel:"Parameters,width=95,markdown=true"`
	Labels     []string `excel:"Labels,width=15"`
	Modifier   []string `excel:"Modifier,width=12"`
	ClassField string   `excel:"Class field,width=10"`
	FilePath   string   `excel:"File path,hide=true"`
}

type Class struct {
	ClassName string
	Extends   []string
}

type TypeConversion struct {
	SourceType      string
	DestinationType string
}

type ReturnTypesDatasetRow struct {
	MethodName string
	TypeLabel  int
}

type MethodGenerationDatasetRow struct {
	ClassName  string
	MethodName string
	Parameters []string
}

type TypeLabel struct {
	Name  string
	Label int
}

type IdealResult struct {
	FilePath              string
	FileType              string
	Identifier            string
	IdentifierType        string
	LineNumber            int
	ColumnNumber          int
	IssueID               string
	IssueAdditionalDetail string
	IssueCategory         string
	IssueDetail           string
	AnalysisDateTime      time.Time
}
