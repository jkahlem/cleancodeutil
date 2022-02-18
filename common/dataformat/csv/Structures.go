package csv

// Generate Marshal / Unmarshal methods (-> Marshaller.go)
//go:generate go run ./marshallerGenerator

type Method struct {
	ClassName  string `excel:"Class name,width=25"`
	ReturnType string `excel:"Return type,width=20"`
	MethodName string `excel:"Method name,width=30"`
	// Parameters are in this format: "<type> <name>" (seperated by a single space)
	Parameters []string `excel:"Parameters,width=95,markdown=true"`
	Exceptions []string `excel:"Exceptions,width=20"`
	Labels     []string `excel:"Labels,width=15"`
	Modifier   []string `excel:"Modifier,width=12"`
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
