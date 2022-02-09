package java

import (
	"encoding/xml"
	"strings"
)

type Method struct {
	XMLName         xml.Name        `xml:"method"`
	MethodName      string          `xml:"name,attr"`
	Annotations     []string        `xml:"annotations>annotation"`
	TypeParameters  []TypeParameter `xml:"typeParameters>typeParameter"`
	ReturnType      Type            `xml:"type"`
	IsChainMethod   bool            `xml:"isChainMethod,attr"`
	MethodNameRange Range           `xml:"methodNameRange>range"`
	ReturnTypeRange Range           `xml:"returnTypeRange>range"`
	Parameters      []Parameter     `xml:"parameters>parameter"`
	parentElement   JavaElement     `xml:"-"`
}

func (method *Method) Path() string {
	if method.Parent() == nil {
		return method.MethodName
	}
	return strings.Join([]string{method.Parent().Path(), method.MethodName}, ".")
}

func (method *Method) Parent() JavaElement {
	return method.parentElement
}
