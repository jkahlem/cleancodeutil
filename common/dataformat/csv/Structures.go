package csv

// Generate Marshal / Unmarshal methods (-> Marshaller.go)
//go:generate go run ./marshallerGenerator

type Method struct {
	MethodName string `excel:"Method name"`
	ReturnType string `excel:"Return type"`
	// Parameters are in this format: "<type> <name>" (seperated by a single space)
	Parameters []string `excel:"Parameters"`
	ClassName  string   `excel:"Class name"`
	Exceptions []string `excel:"Exceptions"`
	Labels     []string `excel:"Labels"`
	Modifier   []string `excel:"Modifier"`
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
	Prefix     string
	MethodName string
	Parameters string
}

type TypeLabel struct {
	Name  string
	Label int
}
