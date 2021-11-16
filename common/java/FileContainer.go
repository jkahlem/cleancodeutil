package java

import "encoding/xml"

type FileContainer interface {
	CodeFiles() []*CodeFile
}

type XMLRoot struct {
	XMLName            xml.Name   `xml:"root"`
	Files              []CodeFile `xml:"files>file"`
	cachedPointerSlice []*CodeFile
}

func (xmlroot *XMLRoot) CodeFiles() []*CodeFile {
	if xmlroot.cachedPointerSlice == nil {
		xmlroot.cachedPointerSlice = make([]*CodeFile, len(xmlroot.Files))
		for i := range xmlroot.Files {
			xmlroot.cachedPointerSlice[i] = &xmlroot.Files[i]
		}
	}
	return xmlroot.cachedPointerSlice
}
