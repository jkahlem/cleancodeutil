package parser

import (
	"regexp"
	"returntypes-langserver/common/utils"
	"strings"
)

type Token struct {
	Range   Range
	Content string
}

func (t Token) IsComment() bool {
	return t.IsLineComment() || t.IsMultilineComment()
}

func (t Token) IsMultilineComment() bool {
	return strings.HasPrefix(t.Content, "/*") && strings.HasSuffix(t.Content, "*/")
}

func (t Token) IsLineComment() bool {
	return strings.HasPrefix(t.Content, "//")
}

func (t Token) IsString() bool {
	return t.IsDoubleQuotedString() || t.IsSingleQuotedString()
}

func (t Token) IsDoubleQuotedString() bool {
	return strings.HasPrefix(t.Content, `"`) && strings.HasSuffix(t.Content, `"`)
}

func (t Token) IsSingleQuotedString() bool {
	return strings.HasPrefix(t.Content, `'`) && strings.HasSuffix(t.Content, `'`)
}

func (t Token) IsAnnotation() bool {
	return strings.HasPrefix(t.Content, "@")
}

func (t Token) IsMethodModifier() bool {
	if t.IsAnnotation() {
		return true
	}
	return utils.StringIsAnyOf(t.Content, "public", "protected", "private", "abstract", "static", "final", "synchronized", "native", "strictfp")
}

func (t Token) IsValid() bool {
	return t.Content != "" && t.Range.Start < t.Range.End
}

type Range struct {
	Start int
	End   int
}

var TokenPattern = regexp.MustCompile(`"(\\"|.)*?"|'(\\'|.)*?'|//.*|/\\*(.|\n)*?\\*/|@?[a-zA-Z][a-zA-Z0-9]*|[{}();<>]`)

type Tokenizer struct {
	input  []byte
	offset int
	token  Token
}

func NewTokenizer(code string) *Tokenizer {
	return &Tokenizer{
		input:  []byte(code),
		offset: 0,
	}
}

func (t *Tokenizer) HasNext() bool {
	if len(t.input) == 0 {
		return false
	}
	match := TokenPattern.FindIndex(t.input)
	if match == nil {
		t.input = nil
		return false
	}
	start, end := match[0], match[1]
	t.token = Token{
		Range: Range{
			Start: t.offset + start,
			End:   t.offset + end,
		},
		Content: string(t.input[start:end]),
	}
	t.offset += end
	t.input = t.input[end:]
	return true
}

func (t *Tokenizer) Token() Token {
	return t.token
}
