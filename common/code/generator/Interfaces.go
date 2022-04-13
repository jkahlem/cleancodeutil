package generator

import "go/ast"

type Interface struct {
	Base
	// The methods of the interface. Does not contain embedded interfaces.
	Methods []InterfaceMethod
}

type InterfaceMethod struct {
	Base
	// The type of the method.
	Type Type
}

// Parses a go file and extracts the interface definitions contained in this file.
// Use parser.ParseComments to include documentations / comments.
func (ctx *context) ParseInterfaces() []Interface {
	interfaces := make([]Interface, 0, 1)

	for i, file := range ctx.files {
		ast.Inspect(file.FileNode, func(n ast.Node) bool {
			if n == nil {
				return false
			} else if typeSpec, interfaceType, ok := ctx.getInterfaceNode(n); ok {
				interfaces = append(interfaces, ctx.createInterface(typeSpec, interfaceType, &ctx.files[i]))
			}
			return true
		})
	}
	return interfaces
}

func (ctx *context) getInterfaceNode(node ast.Node) (*ast.TypeSpec, *ast.InterfaceType, bool) {
	if typeSpec, ok := node.(*ast.TypeSpec); ok {
		if interfaceType, ok := typeSpec.Type.(*ast.InterfaceType); ok {
			return typeSpec, interfaceType, ok
		}
	}
	return nil, nil, false
}

func (ctx *context) createInterface(typeSpec *ast.TypeSpec, srcInterface *ast.InterfaceType, file *SourceFilePair) Interface {
	destInterface := Interface{
		Base:    getBaseValuesFromTypeSpec(typeSpec),
		Methods: make([]InterfaceMethod, 0, len(srcInterface.Methods.List)),
	}
	for _, method := range srcInterface.Methods.List {
		destInterface.Methods = append(destInterface.Methods, ctx.createInterfaceMethods(method, file)...)
	}
	return destInterface
}

func (ctx *context) createInterfaceMethods(srcField *ast.Field, file *SourceFilePair) []InterfaceMethod {
	methods := make([]InterfaceMethod, 0, len(srcField.Names))
	if len(srcField.Names) > 0 {
		for i := range srcField.Names {
			methods = append(methods, ctx.createInterfaceMethod(srcField, i, file))
		}
	}
	return methods
}

func (ctx *context) createInterfaceMethod(srcField *ast.Field, index int, file *SourceFilePair) InterfaceMethod {
	return InterfaceMethod{
		Base: getBaseValuesFromField(srcField, index),
		Type: ctx.ofType(srcField.Type, file),
	}
}
