package generator

import (
	"go/ast"
)

type Struct struct {
	// The name of the struct.
	Name string
	// Documentation above the struct definition (<- like this)
	Documentation string
	// Comments on the same line as the struct definition (= where the word "type" is)
	LineComment string // like this
	// The fields of the struct.
	// Does not contain the fields of an embedded struct, but the existence of the embedded struct (the field has no name then).
	// This is, because the embedded struct might be declared somewhere else (different file/package) and currently only the file scope is checked.
	Fields []StructField
}

type StructField struct {
	// The name of the struct field
	Name string
	// The tag for the struct field
	Tag string
	// Documentation above the field definition (<- like this)
	Documentation string
	// Comments on the same line as the struct field definition
	LineComment string // like this
	// Use *Type() methods on struct field to get the desired type (if it exists and is supported)
	Type Type
}

// Parses a go file and extracts the struct informations contained in this file.
// Use parser.ParseComments to include documentations / comments.
func (ctx *context) ParseStructs() []Struct {
	structs := make([]Struct, 0, 1)

	for i, file := range ctx.files {
		ctx.currentFile = &ctx.files[i]
		ast.Inspect(file.FileNode, func(n ast.Node) bool {
			if n == nil {
				return false
			} else if typeSpec, structType, ok := ctx.getStructNode(n); ok {
				structs = append(structs, ctx.buildStruct(typeSpec, structType))
			}
			return true
		})
	}
	return structs
}

func (ctx *context) getStructNode(node ast.Node) (*ast.TypeSpec, *ast.StructType, bool) {
	if typeSpec, ok := node.(*ast.TypeSpec); ok {
		if structType, ok := typeSpec.Type.(*ast.StructType); ok {
			return typeSpec, structType, ok
		}
	}
	return nil, nil, false
}

func (ctx *context) buildStruct(typeSpec *ast.TypeSpec, srcStruct *ast.StructType) Struct {
	destStruct := Struct{}
	if typeSpec.Name != nil {
		destStruct.Name = typeSpec.Name.Name
	}
	if typeSpec.Doc != nil {
		destStruct.Documentation = typeSpec.Doc.Text()
	}
	if typeSpec.Comment != nil {
		destStruct.LineComment = typeSpec.Comment.Text()
	}
	destStruct.Fields = make([]StructField, 0, len(srcStruct.Fields.List))
	for _, field := range srcStruct.Fields.List {
		destStruct.Fields = append(destStruct.Fields, ctx.buildStructFields(field)...)
	}
	return destStruct
}

// Builds fields from a field definition. Returns a slice, as one field definition might define multiple fields. Example:
//   struct Test {
//	   field1, field2 string
//   }
func (ctx *context) buildStructFields(srcField *ast.Field) []StructField {
	fields := make([]StructField, 0, len(srcField.Names))
	if len(srcField.Names) == 0 {
		// Embedded field
		fields = append(fields, ctx.buildStructField(srcField, -1))
	} else {
		for i := range srcField.Names {
			fields = append(fields, ctx.buildStructField(srcField, i))
		}
	}
	return fields
}

// Builds a single struct field. If index is lower than 0, then the field has no name (so is for example embedding another struct)
func (ctx *context) buildStructField(srcField *ast.Field, index int) StructField {
	field := StructField{}
	if index >= 0 {
		field.Name = srcField.Names[index].Name
	}
	if srcField.Doc != nil {
		field.Documentation = srcField.Doc.Text()
	}
	if srcField.Comment != nil {
		field.LineComment = srcField.Comment.Text()
	}
	if srcField.Tag != nil {
		rawTag := srcField.Tag.Value
		if len(rawTag) > 1 {
			// strip the beginning/ending ` from the tags
			field.Tag = rawTag[1 : len(rawTag)-1]
		}
	}
	field.Type = ctx.ofType(srcField.Type)
	return field
}
