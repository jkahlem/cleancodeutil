package generator

import (
	"fmt"
	"go/ast"
	"go/token"
)

type Const struct {
	Name  string
	Value Expr
}

func (ctx *context) ParseConsts() []Const {
	consts := make([]Const, 0, 1)

	for i, file := range ctx.files {
		ast.Inspect(file.FileNode, func(node ast.Node) bool {
			if node == nil {
				return false
			} else if declNode, ok := node.(*ast.GenDecl); ok && declNode.Tok == token.CONST {
				consts = append(consts, ctx.getConstNodesFromSpec(declNode.Specs, &ctx.files[i])...)
			}
			return true
		})
	}
	return consts
}

func (ctx *context) getConstNodesFromSpec(specs []ast.Spec, file *SourceFilePair) []Const {
	consts := make([]Const, 0, len(specs))
	for _, s := range specs {
		if spec, ok := s.(*ast.ValueSpec); ok {
			consts = append(consts, ctx.getConstNodes(spec.Names, spec.Values, file)...)
		}
	}
	return consts
}

func (ctx *context) getConstNodes(names []*ast.Ident, values []ast.Expr, file *SourceFilePair) []Const {
	if len(names) != len(values) {
		panic(fmt.Errorf("Got const declaration with %d identifiers but %d value expression.", len(names), len(values)))
	}
	consts := make([]Const, len(names))
	for i := range names {
		consts[i] = Const{
			Name:  names[i].Name,
			Value: ctx.ofExpr(values[i], file),
		}
	}
	return consts
}

func (ctx *context) getGenDeclNode(node ast.Node) (*ast.TypeSpec, *ast.StructType, bool) {
	if typeSpec, ok := node.(*ast.TypeSpec); ok {
		if structType, ok := typeSpec.Type.(*ast.StructType); ok {
			return typeSpec, structType, ok
		}
	}
	return nil, nil, false
}
