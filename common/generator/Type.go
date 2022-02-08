package generator

import "go/ast"

type Type struct {
	base ast.Expr
	file *SourceFilePair
	ctx  *context
}

func (ctx *context) ofType(expr ast.Expr) Type {
	return Type{
		base: expr,
		file: ctx.currentFile,
		ctx:  ctx,
	}
}

// Returns the source code snippet for the type declaration
func (t *Type) Code() string {
	start, end := t.base.Pos()-1, t.base.End()-1
	if t.file == nil || len(t.file.Source) < int(end) {
		return ""
	}
	return t.file.Source[start:end]
}

// Builds the given type as a function type if it is a function type (ok will be true). Otherwise, ok will be false.
func (t *Type) FunctionType() (fnType FunctionType, ok bool) {
	if funcType, ok := t.base.(*ast.FuncType); ok && t.ctx != nil {
		return t.ctx.buildFunctionType(funcType), true
	}
	return FunctionType{}, false
}
