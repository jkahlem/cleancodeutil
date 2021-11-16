package java

import (
	"encoding/xml"
	"strings"
)

type Import struct {
	XMLName       xml.Name  `xml:"import"`
	ImportPath    string    `xml:",chardata"`
	IsWildcard    bool      `xml:"isWildcard,attr"`
	IsStatic      bool      `xml:"isStatic,attr"`
	parentElement *CodeFile `xml:"-"`
}

func (_import *Import) Path() string {
	if _import.Parent() == nil {
		return _import.ImportPath
	}
	return strings.Join([]string{_import.Parent().Path(), _import.ImportPath}, ".")
}

func (_import *Import) Parent() JavaElement {
	return _import.parentElement
}
