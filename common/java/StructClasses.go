package java

import (
	"encoding/xml"
	"strings"

	"returntypes-langserver/common/errors"
	"returntypes-langserver/common/packagetree"
)

const JavaErrorTitle = "Error"

// Class Types
const (
	ENUM      = "ENUM"
	INTERFACE = "INTERFACE"
	CLASS     = "CLASS"
)

type Class struct {
	XMLName           xml.Name        `xml:"class"`
	ClassName         string          `xml:"name,attr"`
	ClassType         string          `xml:"type,attr"`
	Modifiers         []string        `xml:"modifiers>modifier"`
	Classes           []*Class        `xml:"classes>class"`
	Methods           []Method        `xml:"methods>method"`
	TypeParameters    []TypeParameter `xml:"typeParameters>typeParameter"`
	ExtendsImplements []Type          `xml:"extends>type"`
	parentElement     JavaElement     `xml:"-"`
}

// Finds a subclass with the specified name.
func (class *Class) FindChild(name string, options packagetree.SelectionOptions) packagetree.Node {
	for i, subClass := range class.Classes {
		if subClass.ClassName == name {
			return class.Classes[i]
		}
	}
	return nil
}

// Adds a subclass to the class.
func (class *Class) AddChild(node packagetree.Node, options packagetree.SelectionOptions) errors.Error {
	switch t := node.(type) {
	case *Class:
		class.Classes = append(class.Classes, t)
		t.parentElement = class
		packagetree.ForSubscribableParents(class, func(subscribable packagetree.Subscribable) {
			subscribable.OnChildAdded(t)
		})
		return nil
	}
	return errors.New(JavaErrorTitle, "Only classes supported as children of class nodes")
}

// Removes a subclass.
func (class *Class) RemoveChild(name string, options packagetree.SelectionOptions) errors.Error {
	for i, subClass := range class.Classes {
		if subClass.TargetNode(name, options) != nil {
			class.Classes[i].parentElement = nil
			if i+1 < len(class.Classes) {
				// if not last element, copy all following elements to one field before
				copy(class.Classes[i:], class.Classes[i+1:])
			}
			// shorten slice by 1 element
			class.Classes = class.Classes[:len(class.Classes)]
			return nil
		}
	}
	return errors.New(JavaErrorTitle, "No element to remove")
}

// If the name is equal to the class name, returns a pointer to this class. If options.FindOnlyPublicClass is true,
// returns nil if this class is not public.
func (class *Class) TargetNode(name string, options packagetree.SelectionOptions) packagetree.Node {
	if class.ClassName != name || options.FindOnlyPublicClasses && !class.IsPublicClass() {
		return nil
	}
	return class
}

// Sets the parent of this node.
func (class *Class) SetParentNode(parent packagetree.Node) {
	switch t := parent.(type) {
	case JavaElement:
		class.parentElement = t
	}
}

// Returns the parent of this node.
func (class *Class) ParentNode() packagetree.Node {
	if class.parentElement == nil {
		return nil
	}

	switch t := class.parentElement.(type) {
	case *Class:
		return t
	case *CodeFile:
		return t
	}
	return nil
}

// Returns true if the class has a public modifier.
func (class *Class) IsPublicClass() bool {
	if class.Modifiers == nil {
		return false
	}
	for _, modifier := range class.Modifiers {
		if modifier == "public" {
			return true
		}
	}
	return false
}

// Returns all methods inside the class definition of the file.
// This means, all methods inherited from extended classes are not returned.
func (class *Class) GetAllMethodsInSameFile() []*Method {
	methods := make([]*Method, len(class.Methods))
	for i := range class.Methods {
		methods[i] = &class.Methods[i]
	}

	for i := range class.Classes {
		if class.Classes[i] == nil {
			continue
		}

		subMethods := class.Classes[i].GetAllMethodsInSameFile()
		if len(subMethods) > 0 {
			methods = append(methods, subMethods...)
		}
	}
	return methods
}

func (class *Class) Path() string {
	if class.Parent() == nil {
		return class.ClassName
	}
	return strings.Join([]string{class.Parent().Path(), class.ClassName}, ".")
}

func (class *Class) Parent() JavaElement {
	return class.parentElement
}

func (class *Class) NodeName() string {
	return class.ClassName
}
