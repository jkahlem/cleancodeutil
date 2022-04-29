package csv

// Generate Marshal / Unmarshal methods (-> Marshaller.go)
//go:generate go run ./marshallerGenerator

type Method struct {
	ClassName  string   `excel:"Class name,width=25"`
	ReturnType string   `excel:"Return type,width=20"`
	MethodName string   `excel:"Method name,width=30"`
	Parameters []string `excel:"Parameters,width=95"`
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
	ClassName    string
	MethodName   string
	ReturnType   string
	IsStatic     bool
	Parameters   []string
	ContextTypes []string
}

type TypeLabel struct {
	Name  string
	Label int
}

type FileContextTypes struct {
	FilePath     string
	ContextTypes []string
}
