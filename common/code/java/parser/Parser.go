// This package has some (more naive/simple) functionalities to parse java code which are needed in this project.
// This is mainly just implemented to parse specific information from java code which contains parser errors (e.g.
// code which is incomplete as it is still edited) which the crawler (/ javaparser library) is currently not
// performing as required.
package parser

type Method struct {
	Name        Token
	Type        Token
	RoundBraces Token
	Annotations []Token
	IsStatic    bool
}

type Class struct {
	ClassType string
	Name      Token
	Methods   []Method
	Classes   []Class
}

const (
	ClassContext     = "class"
	InterfaceContext = "interface"
	EnumContext      = "enum"
)

// Returns an array of method definitions
func Parse(code string) *Class {
	tokenizer := NewTokenizer(code)
	statement := make([]Token, 0, 10)
	context := NewClassStack()
	var topLevelClass *Class
	afterAnnotation := false
	for tokenizer.HasNext() {
		token := tokenizer.Token()
		if token.IsComment() || token.IsString() {
			continue
		}

		if token.IsAnnotation() {
			statement = append(statement, token)
			afterAnnotation = true
			continue
		} else if afterAnnotation && token.Content == "(" {
			skipBlock(tokenizer, "(", ")", 1)
			afterAnnotation = false
			continue
		}
		afterAnnotation = false

		if token.Content == ";" {
			if top, ok := context.Peek(); ok && top.ClassType == InterfaceContext {
				// statement is method definition
				top.Methods = append(top.Methods, getMethodFromStatement(statement, code))
			}
			statement = statement[:0]
		} else if token.Content == "{" {
			if class := getClassFromStatement(statement); class != nil {
				if topLevelClass == nil {
					topLevelClass = class
				}
				context.Push(class)
			}
			statement = statement[:0]
		} else if token.Content == "}" {
			if poppedClass, ok := context.Pop(); ok {
				if top, ok := context.Peek(); ok {
					top.Classes = append(top.Classes, *poppedClass)
				}
			}
			statement = statement[:0]
		} else if token.Content == ")" {
			statement = append(statement, token)
			if top, ok := context.Peek(); ok && top.ClassType == ClassContext {
				// statement is method definition
				top.Methods = append(top.Methods, getMethodFromStatement(statement, code))
				statement = statement[:0]

				// skip until leaving method
				// If the method has no curly braces (as it might be still written; this is a parser error but we need at least the above information) then
				// a method/class which follows this method definition will also be skipped.
				// This is more complicated to fix (as method definitions might also be followed by the throws declarations etc.),
				// but as (probably) only the written method is needed, it is atm not worth the effort.
				skipBlock(tokenizer, "{", "}", 0)
			}
		} else {
			statement = append(statement, token)
		}
	}

	return topLevelClass
}

func getClassFromStatement(statement []Token) *Class {
	class := Class{}
	for i, t := range statement {
		switch t.Content {
		case "class":
			class.ClassType = ClassContext
		case "interface":
			class.ClassType = InterfaceContext
		case "enum":
			class.ClassType = EnumContext
		default:
			continue
		}
		if i+1 == len(statement) {
			return nil
		}
		class.Name = statement[i+1]
	}
	return &class
}

func getMethodFromStatement(statement []Token, code string) Method {
	method := Method{}
	parseParameterList := false
	for i, t := range statement {
		if t.Content == "(" {
			parseParameterList = true
			if i == 0 {
				return Method{}
			}
			method.Name = statement[i-1]
			method.RoundBraces.Range.Start = t.Range.Start
		} else if t.Content == ")" {
			method.RoundBraces.Range.End = t.Range.End
		} else if !parseParameterList {
			if t.IsAnnotation() {
				method.Annotations = append(method.Annotations, t)
			} else if t.Content == "static" {
				method.IsStatic = true
			}
		}
	}
	if method.RoundBraces.Range.End < method.RoundBraces.Range.Start {
		// set the end to the last part of the statement
		method.RoundBraces.Range.End = statement[len(statement)-1].Range.End
	}
	r := method.RoundBraces.Range
	method.RoundBraces.Content = code[r.Start:r.End]
	return method
}

func containsToken(tokens []Token, str string) bool {
	for _, t := range tokens {
		if t.Content == str {
			return true
		}
	}
	return false
}

func joinTokens(tokens []Token) string {
	output := ""
	for i, t := range tokens {
		if i > 0 {
			output += " " + t.Content
		} else {
			output += t.Content
		}
	}
	return output
}

// Skips the block beginning with the pattern in and ending on the pattern out, which is nestable (like function blocks in curly braces and so on)
// The level parameter sets on which nesting level of the block the tokenizer starts. The skipping continues until this level gets 0.
// If level = 0 is passed, then everything is skipped until the block begins and ends.
func skipBlock(tokenizer *Tokenizer, in, out string, level int) {
	for tokenizer.HasNext() {
		str := tokenizer.Token().Content
		if str == in {
			level++
		} else if str == out {
			level--
			if level <= 0 {
				return
			}
		}
	}
}
