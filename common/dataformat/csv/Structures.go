package csv

// Generate Marshal / Unmarshal methods (-> Marshaller.go)
//go:generate go run ./marshallerGenerator

type Method struct {
	MethodName string
	ReturnType string
	Labels     []string
	FilePath   string
	// Parameters are in this format: "<type> <name>" (seperated by a single space)
	Parameters []string
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
	Prefix     string
	MethodName string
	Parameters string
}

type TypeLabel struct {
	Name  string
	Label int
}
