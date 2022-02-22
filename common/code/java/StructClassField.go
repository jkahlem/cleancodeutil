package java

import (
	"encoding/xml"
	"strings"
)

type ClassField struct {
	XMLName       xml.Name    `xml:"field"`
	Name          string      `xml:"name"`
	Type          Type        `xml:"type"`
	parentElement JavaElement `xml:"-"`
}

func (field *ClassField) Path() string {
	if field.Parent() == nil {
		return field.Name
	}
	return strings.Join([]string{field.Parent().Path(), field.Name}, ".")
}

func (field *ClassField) Parent() JavaElement {
	return field.parentElement
}
