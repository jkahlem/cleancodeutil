package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const ValidExampleCode = `package com.example;
/**/
/* "multi
line*/
public class SomeClass { // valid line comment()
	private String name;

	@Override
	public String getName() {
		return name;
	}

	public void doSomething
		(String str, int value) {
		if (name == "some /* \\" name()") {
			System.out.println("This is valid code.");
		}
	}
}`

func TestTokenizerWithValidCode(t *testing.T) {

	// when
	tokenized := getTokens(ValidExampleCode)

	// then
	assert.Equal(t, `package,com,example,;,/**/,/* "multi
line*/,public,class,SomeClass,{,// valid line comment(),private,String,name,;,public,String,getName,(,),{,return,name,;,},public,void,doSomething,(,String,str,int,value,),{,if,(,name,"some /* \\" name()",),{,System,out,println,(,"This is valid code.",),;,},},}`, strings.Join(tokenized, ","))
}

func TestParseClass(t *testing.T) {
	// when
	class := Parse(ValidExampleCode)
	methods := class.Methods

	// then
	assert.Len(t, methods, 2)
	assert.Equal(t, "SomeClass", class.Name.Content)
	assert.Equal(t, ClassContext, class.ClassType)
	assert.Equal(t, "getName", methods[0].Name.Content)
	assert.Equal(t, "()", methods[0].RoundBraces.Content)
	assert.Equal(t, "@Override", methods[0].Annotations[0].Content)
	assert.Equal(t, "doSomething", methods[1].Name.Content)
	assert.Equal(t, "(String str, int value)", methods[1].RoundBraces.Content)
}

func TestParseInterface(t *testing.T) {
	// when
	class := Parse(`
public interface ExampleInterface {
	public void doSomething(String str);
}`)
	methods := class.Methods

	// then
	assert.Len(t, methods, 1)
	assert.Equal(t, "ExampleInterface", class.Name.Content)
	assert.Equal(t, InterfaceContext, class.ClassType)
	assert.Equal(t, "doSomething", methods[0].Name.Content)
	assert.Equal(t, "(String str)", methods[0].RoundBraces.Content)
}

func getTokens(code string) []string {
	tokenizer := NewTokenizer(code)
	tokens := make([]string, 0, 64)

	for tokenizer.HasNext() {
		tokens = append(tokens, tokenizer.Token().Content)
	}
	return tokens
}
