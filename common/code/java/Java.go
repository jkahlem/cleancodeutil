// This package contains structures representing java code elements and utilities like the resolver
// to work with these elements.
package java

import (
	"encoding/xml"
)

type JavaElement interface {
	Accept(visitor Visitor)
	Path() string
	Parent() JavaElement
}

type Range struct {
	XMLName xml.Name `xml:"range"`
	Begin   Position `xml:"begin"`
	End     Position `xml:"end"`
}

type Position struct {
	Line int `xml:"line,attr"`
	Col  int `xml:"col,attr"`
}
