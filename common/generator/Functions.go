package generator

import "go/ast"

type FunctionType struct {
	In  []Parameter
	Out []Parameter
}

type Parameter struct {
	Name string
	Type Type
}

func (ctx *context) buildFunctionType(srcType *ast.FuncType) FunctionType {
	destType := FunctionType{
		In:  ctx.buildParametersFromList(srcType.Params),
		Out: ctx.buildParametersFromList(srcType.Results),
	}
	return destType
}

func (ctx *context) buildParametersFromList(srcList *ast.FieldList) []Parameter {
	if srcList == nil {
		return nil
	}
	destList := make([]Parameter, 0, len(srcList.List))
	for _, par := range srcList.List {
		destList = append(destList, ctx.buildParameters(par)...)
	}
	return destList
}

// Builds parameters from a field declaration belonging to a function type.
// Returns a list of parameters, as one field declaration might declare multiple parameters, e.g. func (par1, par2 string)
func (ctx *context) buildParameters(srcPar *ast.Field) []Parameter {
	parameters := make([]Parameter, 0, len(srcPar.Names))
	if len(srcPar.Names) == 0 {
		// Unnamed parameter (e.g. func (string))
		parameters = append(parameters, ctx.buildParameter(srcPar, -1))
	} else {
		for index := range srcPar.Names {
			parameters = append(parameters, ctx.buildParameter(srcPar, index))
		}
	}
	return parameters
}

// Builds a single parameter with the given index. index might be lower than one to indicate, that it is an unnamed parameter.
func (ctx *context) buildParameter(srcPar *ast.Field, index int) Parameter {
	destPar := Parameter{}
	if index >= 0 {
		destPar.Name = srcPar.Names[index].Name
	}
	destPar.Type = ctx.ofType(srcPar.Type)
	return destPar
}
