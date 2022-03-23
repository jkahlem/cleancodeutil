// This package has some (more naive/simple) functionalities to parse java code which are needed in this project.
// This is mainly just implemented to parse specific information from java code which contains parser errors (e.g.
// code which is incomplete as it is still edited) which the crawler (/ javaparser library) is currently not
// performing as required.
package parser

import (
	"returntypes-langserver/common/utils"
)

type Method struct {
	Name        Token
	Type        Token
	RoundBraces Token
}

const (
	ClassContext     = "class"
	InterfaceContext = "interface"
	EnumContext      = "enum"
)

// Returns an array of method definitions
func ParseMethods(code string) []Method {
	tokenizer := NewTokenizer(code)
	statement := make([]Token, 0, 10)
	context := utils.NewStringStack()
	methods := make([]Method, 0, 8)
	afterAnnotation := false
	for tokenizer.HasNext() {
		token := tokenizer.Token()
		if token.IsComment() || token.IsString() {
			continue
		}

		if token.IsAnnotation() {
			afterAnnotation = true
			continue
		} else if afterAnnotation && token.Content == "(" {
			skipBlock(tokenizer, "(", ")")
			afterAnnotation = false
			continue
		}
		afterAnnotation = false

		if token.Content == ";" {
			if top, ok := context.Peek(); ok && top == InterfaceContext {
				// statement is method definition
				methods = append(methods, getMethodFromStatement(statement))
			}
			statement = statement[:0]
		} else if token.Content == "{" {
			contextChanged := false
			for _, t := range statement {
				switch t.Content {
				case "class":
					context.Push(ClassContext)
					contextChanged = true
				case "interface":
					context.Push(InterfaceContext)
					contextChanged = true
				case "enum":
					context.Push(EnumContext)
					contextChanged = true
				}
				if contextChanged {
					break
				}
			}
			if !contextChanged {
				if top, ok := context.Peek(); ok && top == ClassContext {
					// statement is method definition
					methods = append(methods, getMethodFromStatement(statement))

					// skip until leaving method (without current skip until implementation ...)
					skipBlock(tokenizer, "{", "}")
				}
			}
			statement = statement[:0]
		} else if token.Content == "}" {
			context.Pop()
			statement = statement[:0]
		} else {
			statement = append(statement, token)
		}
	}

	for i := range methods {
		r := methods[i].RoundBraces.Range
		methods[i].RoundBraces.Content = code[r.Start:r.End]
	}
	return methods
}

func getMethodFromStatement(statement []Token) Method {
	method := Method{}
	for i, t := range statement {
		if t.Content == "(" {
			if i == 0 {
				return Method{}
			}
			method.Name = statement[i-1]
			method.RoundBraces.Range.Start = t.Range.Start
		} else if t.Content == ")" {
			method.RoundBraces.Range.End = t.Range.End
		}
	}
	if method.RoundBraces.Range.End < method.RoundBraces.Range.Start {
		// set the end to the last part of the statement
		method.RoundBraces.Range.End = statement[len(statement)-1].Range.End
	}
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
// The tokenizer is expected to be actually on the first level of the block (one occurence of in is already read)
func skipBlock(tokenizer *Tokenizer, in, out string) {
	level := 1
	for tokenizer.HasNext() {
		str := tokenizer.Token().Content
		if str == in {
			level++
		} else if str == out {
			level--
			if level == 0 {
				return
			}
		}
	}
}
