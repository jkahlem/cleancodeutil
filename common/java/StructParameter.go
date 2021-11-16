package java

import (
	"encoding/xml"
	"strings"
)

type Parameter struct {
	XMLName       xml.Name    `xml:"parameter"`
	Name          string      `xml:"name,attr"`
	Type          Type        `xml:"type"`
	parentElement JavaElement `xml:"-"`
}

func (par *Parameter) Path() string {
	if par.Parent() == nil {
		return par.Name
	}
	return strings.Join([]string{par.Parent().Path(), par.Name}, ".")
}

func (par *Parameter) Parent() JavaElement {
	return par.parentElement
}
