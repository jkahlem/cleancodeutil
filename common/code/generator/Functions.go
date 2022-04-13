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

func (ctx *context) createFunctionType(srcType *ast.FuncType, file *SourceFilePair) FunctionType {
	destType := FunctionType{
		In:  ctx.createParametersFromList(srcType.Params, file),
		Out: ctx.createParametersFromList(srcType.Results, file),
	}
	return destType
}

func (ctx *context) createParametersFromList(srcList *ast.FieldList, file *SourceFilePair) []Parameter {
	if srcList == nil {
		return nil
	}
	destList := make([]Parameter, 0, len(srcList.List))
	for _, par := range srcList.List {
		destList = append(destList, ctx.createParameters(par, file)...)
	}
	return destList
}

// Creates parameters from a field declaration belonging to a function type.
// Returns a list of parameters, as one field declaration might declare multiple parameters, e.g. func (par1, par2 string)
func (ctx *context) createParameters(srcPar *ast.Field, file *SourceFilePair) []Parameter {
	parameters := make([]Parameter, 0, len(srcPar.Names))
	if len(srcPar.Names) == 0 {
		// Unnamed parameter (e.g. func (string))
		parameters = append(parameters, ctx.createParameter(srcPar, -1, file))
	} else {
		for index := range srcPar.Names {
			parameters = append(parameters, ctx.createParameter(srcPar, index, file))
		}
	}
	return parameters
}

// Creates a single parameter with the given index. index might be lower than one to indicate, that it is an unnamed parameter.
func (ctx *context) createParameter(srcPar *ast.Field, index int, file *SourceFilePair) Parameter {
	destPar := Parameter{}
	if index >= 0 {
		destPar.Name = srcPar.Names[index].Name
	}
	destPar.Type = ctx.ofType(srcPar.Type, file)
	return destPar
}

type FunctionDeclaration struct {
	Name          string
	Documentation string
	ReceiverType  string
	Type          Type
}

// Parses a go file and extracts the function declarations contained in this file.
// Use parser.ParseComments to include documentations / comments.
func (ctx *context) ParseFunctionDeclarations() []FunctionDeclaration {
	declarations := make([]FunctionDeclaration, 0, 1)

	for i, file := range ctx.files {
		ast.Inspect(file.FileNode, func(n ast.Node) bool {
			if n == nil {
				return false
			} else if funcDecl, ok := ctx.getFunctionDeclarationNode(n); ok {
				declarations = append(declarations, ctx.createFunctionDeclaration(funcDecl, &ctx.files[i]))
			}
			return true
		})
	}
	return declarations
}

func (ctx *context) getFunctionDeclarationNode(node ast.Node) (*ast.FuncDecl, bool) {
	if funcDecl, ok := node.(*ast.FuncDecl); ok {
		return funcDecl, ok
	}
	return nil, false
}

func (ctx *context) createFunctionDeclaration(srcFuncDecl *ast.FuncDecl, file *SourceFilePair) FunctionDeclaration {
	destDecl := FunctionDeclaration{
		Type: ctx.ofType(srcFuncDecl.Type, file),
	}
	if srcFuncDecl.Name != nil {
		destDecl.Name = srcFuncDecl.Name.Name
	}
	if srcFuncDecl.Doc != nil {
		destDecl.Documentation = srcFuncDecl.Doc.Text()
	}
	if srcFuncDecl.Recv != nil && len(srcFuncDecl.Recv.List) == 1 {
		t := ctx.ofType(srcFuncDecl.Recv.List[0].Type, file)
		destDecl.ReceiverType = t.Code()
	}

	return destDecl
}

func (ctx *context) getReceivers(receiverList *ast.FieldList) []string {
	if receiverList == nil {
		return nil
	}
	receivers := make([]string, 0, 1)
	for _, receiver := range receiverList.List {
		receivers = append(receivers, receiver.Names[0].Name)
	}
	return nil
}
