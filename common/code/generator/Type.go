package generator

import (
	"go/ast"
)

type Type struct {
	base   ast.Expr
	source *SourceFilePair
	ctx    *context
}

func (ctx *context) ofType(expr ast.Expr, file *SourceFilePair) Type {
	return Type{
		base:   expr,
		source: file,
		ctx:    ctx,
	}
}

// Returns the content of the original source code used to declare this type.
func (t *Type) Code() string {
	start, end := t.ctx.fileset.Position(t.base.Pos()), t.ctx.fileset.Position(t.base.End())
	if len(t.source.Source) < int(end.Offset) {
		return ""
	}
	return t.source.Source[start.Offset:end.Offset]
}

// Builds the given type as a function type if it is a function type (ok will be true). Otherwise, ok will be false.
func (t *Type) FunctionType() (fnType FunctionType, ok bool) {
	if funcType, ok := t.base.(*ast.FuncType); ok && t.ctx != nil {
		return t.ctx.createFunctionType(funcType, t.source), true
	}
	return FunctionType{}, false
}
