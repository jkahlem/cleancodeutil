package java

// Visitor for making parent and child references in the java structures.
type ConnectorVisitor struct{}

// Connects imports/classes to the code file.
func (visitor *ConnectorVisitor) VisitCodeFile(codeFile *CodeFile) {
	for i := range codeFile.Imports {
		codeFile.Imports[i].parentElement = codeFile
	}
	for i := range codeFile.Classes {
		if codeFile.Classes[i] == nil {
			continue
		}

		codeFile.Classes[i].parentElement = codeFile
		codeFile.Classes[i].Accept(visitor)
	}
}

// Connects extended classes, methods, fields, type parameters and sub classes to it's parent class.
func (visitor *ConnectorVisitor) VisitClass(class *Class) {
	for i := range class.ExtendsImplements {
		class.ExtendsImplements[i].parentElement = class
	}
	for i := range class.Methods {
		class.Methods[i].parentElement = class
		class.Methods[i].Accept(visitor)
	}
	for i := range class.TypeParameters {
		class.TypeParameters[i].parentElement = class
		class.TypeParameters[i].Accept(visitor)
	}
	for i := range class.Classes {
		if class.Classes[i] == nil {
			continue
		}

		class.Classes[i].parentElement = class
		class.Classes[i].Accept(visitor)
	}
}

func (visitor *ConnectorVisitor) VisitImport(*Import) {
	// Do nothing. Just needed for implementing the Visitor interface
}

// Connects the return type and type parameter of a method to the method.
func (visitor *ConnectorVisitor) VisitMethod(method *Method) {
	method.ReturnType.parentElement = method

	for i := range method.TypeParameters {
		method.TypeParameters[i].parentElement = method
		method.TypeParameters[i].Accept(visitor)
	}
	for i := range method.Parameters {
		method.Parameters[i].parentElement = method
		method.Parameters[i].Type.parentElement = &method.Parameters[i]
	}
}

// Connects the type bounds of a type parameter to the type parameter.
func (visitor *ConnectorVisitor) VisitTypeParameter(typeParameter *TypeParameter) {
	for i := range typeParameter.TypeBounds {
		typeParameter.TypeBounds[i].parentElement = typeParameter
	}
}
