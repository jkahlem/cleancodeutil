package diagnostics

import (
	"fmt"
	"returntypes-langserver/common/code/java"
	"returntypes-langserver/common/code/packagetree"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/languageserver/lsp"
	"returntypes-langserver/processing/typeclasses"
	"returntypes-langserver/services/predictor"
)

const DiagnosticsErrorTitle = "Error"

type Creator struct {
	mapper          TypeMapper
	typeClassMapper typeclasses.Mapper
	tree            *packagetree.Tree
}

// Scans the given file for unexpected return types and creates diagnostics for them.
func (d *Creator) CreateDiagnosticsForFile(file *java.CodeFile, tree *packagetree.Tree) ([]ExpectedReturnTypeDiagnostic, errors.Error) {
	if file == nil {
		return nil, errors.New(DiagnosticsErrorTitle, "File does not exist")
	} else if tree == nil {
		return nil, errors.New(DiagnosticsErrorTitle, "Package tree not set")
	}
	d.tree = tree

	methods := file.GetAllMethods()
	predictionMappings, err := d.mapper.CreatePredictionMappings(methods)
	if err != nil {
		return nil, err
	}
	diagnostics := d.createDiagnosticsForMethods(methods, predictionMappings)
	return diagnostics, nil
}

// Creates diagnostics for the given methods.
func (d *Creator) createDiagnosticsForMethods(methods []*java.Method, mappings predictor.MethodTypeMap) []ExpectedReturnTypeDiagnostic {
	diagnostics := make([]ExpectedReturnTypeDiagnostic, 0, len(methods))
	for _, method := range methods {
		if method == nil {
			continue
		}

		actualType, expectedType := d.getActualAndExpectedMethodTypes(method, mappings)
		if actualType != expectedType {
			diagnostics = append(diagnostics, d.createDiagnostic(expectedType, method))
		}
	}
	return diagnostics
}

func (d *Creator) createDiagnostic(expectedType string, method *java.Method) ExpectedReturnTypeDiagnostic {
	return ExpectedReturnTypeDiagnostic{
		ExpectedReturnType: expectedType,
		MethodNameRange:    lsp.FromJavaRange(method.MethodNameRange),
		ReturnTypeRange:    lsp.FromJavaRange(method.ReturnTypeRange),
	}
}

func (d *Creator) getActualAndExpectedMethodTypes(method *java.Method, mappings predictor.MethodTypeMap) (actual, expected string) {
	if typeClass, err := d.getTypeClassForMethodReturnType(method); err != nil {
		return "", ""
	} else {
		actual = typeClass
		expected = mappings[predictor.GetPredictableMethodName(method.MethodName)]
		return
	}
}

// Maps the return type of the method to it's type class.
func (d *Creator) getTypeClassForMethodReturnType(method *java.Method) (string, errors.Error) {
	resolvedType, _ := java.Resolve(&method.ReturnType, d.tree)
	mapper, err := d.getTypeClassMapper()
	if err != nil {
		return "", err
	}
	typeClass, err := mapper.MapReturnTypeToTypeClass(resolvedType, java.GetMethodLabels(method))
	if err != nil {
		return "", err
	}
	return typeClass, nil
}

func (d *Creator) getTypeClassMapper() (typeclasses.Mapper, errors.Error) {
	if d.typeClassMapper == nil {
		set, ok := configuration.FindDatasetByReference(configuration.LanguageServerReturntypesDataset())
		if !ok {
			return nil, errors.New(DiagnosticsErrorTitle, fmt.Sprintf("Returntypes validation dataset not found: %s", configuration.LanguageServerReturntypesDataset()))
		}
		return typeclasses.New(d.tree, set.SpecialOptions.TypeClasses)
	} else {
		d.typeClassMapper.SetPackageTree(d.tree)
		return d.typeClassMapper, nil
	}
}
