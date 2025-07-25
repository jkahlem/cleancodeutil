package generator

import (
	"go/ast"
)

type Struct struct {
	Base
	// The fields of the struct.
	// Does not contain the fields of an embedded struct, but the existence of the embedded struct (the field has no name then).
	// This is, because the embedded struct might be declared somewhere else (different file/package) and currently only the file scope is checked.
	Fields []StructField
}

type StructField struct {
	Base
	// The tag for the struct field
	Tag string
	// Use *Type() methods on struct field to get the desired type (if it exists and is supported)
	Type Type
}

// Parses a go file and extracts the struct definitions contained in this file.
// Use parser.ParseComments to include documentations / comments.
func (ctx *context) ParseStructs() []Struct {
	structs := make([]Struct, 0, 1)

	for i, file := range ctx.files {
		ast.Inspect(file.FileNode, func(n ast.Node) bool {
			if n == nil {
				return false
			} else if typeSpec, structType, ok := ctx.getStructNode(n); ok {
				structs = append(structs, ctx.createStruct(typeSpec, structType, &ctx.files[i]))
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

func (ctx *context) createStruct(typeSpec *ast.TypeSpec, srcStruct *ast.StructType, file *SourceFilePair) Struct {
	destStruct := Struct{
		Base:   getBaseValuesFromTypeSpec(typeSpec),
		Fields: make([]StructField, 0, len(srcStruct.Fields.List)),
	}
	for _, field := range srcStruct.Fields.List {
		destStruct.Fields = append(destStruct.Fields, ctx.createStructFields(field, file)...)
	}
	return destStruct
}

// Creates fields from a field definition. Returns a slice, as one field definition might define multiple fields. Example:
//   struct Test {
//	   field1, field2 string
//   }
func (ctx *context) createStructFields(srcField *ast.Field, file *SourceFilePair) []StructField {
	fields := make([]StructField, 0, len(srcField.Names))
	if len(srcField.Names) == 0 {
		// Embedded field
		fields = append(fields, ctx.createStructField(srcField, -1, file))
	} else {
		for i := range srcField.Names {
			fields = append(fields, ctx.createStructField(srcField, i, file))
		}
	}
	return fields
}

// Creates a single struct field. If index is lower than 0, then the field has no name (so is for example embedding another struct)
func (ctx *context) createStructField(srcField *ast.Field, index int, file *SourceFilePair) StructField {
	field := StructField{
		Base: getBaseValuesFromField(srcField, index),
		Type: ctx.ofType(srcField.Type, file),
	}
	if srcField.Tag != nil {
		rawTag := srcField.Tag.Value
		if len(rawTag) > 1 {
			// strip the beginning/ending ` from the tags
			field.Tag = rawTag[1 : len(rawTag)-1]
		}
	}
	return field
}
