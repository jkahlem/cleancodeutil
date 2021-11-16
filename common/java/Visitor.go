package java

type Visitor interface {
	VisitCodeFile(*CodeFile)
	VisitClass(*Class)
	VisitImport(*Import)
	VisitMethod(*Method)
	VisitTypeParameter(*TypeParameter)
}

func (class *Class) Accept(visitor Visitor) {
	visitor.VisitClass(class)
}

func (codeFile *CodeFile) Accept(visitor Visitor) {
	visitor.VisitCodeFile(codeFile)
}

func (_import *Import) Accept(visitor Visitor) {
	visitor.VisitImport(_import)
}

func (method *Method) Accept(visitor Visitor) {
	visitor.VisitMethod(method)
}

func (typeParameter *TypeParameter) Accept(visitor Visitor) {
	visitor.VisitTypeParameter(typeParameter)
}
