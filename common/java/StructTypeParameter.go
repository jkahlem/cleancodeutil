package java

import (
	"encoding/xml"
	"strings"
)

type TypeParameter struct {
	XMLName           xml.Name    `xml:"typeParameter"`
	TypeParameterName string      `xml:"name,attr"`
	TypeBounds        []Type      `xml:"typeBound"`
	parentElement     JavaElement `xml:"-"`
}

func (typeParameter *TypeParameter) Path() string {
	if typeParameter.Parent() == nil {
		return typeParameter.TypeParameterName
	}
	return strings.Join([]string{typeParameter.Parent().Path(), typeParameter.TypeParameterName}, ".")
}

func (typeParameter *TypeParameter) Parent() JavaElement {
	return typeParameter.parentElement
}
