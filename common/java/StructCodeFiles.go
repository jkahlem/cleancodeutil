package java

import (
	"encoding/xml"
	"fmt"

	"returntypes-langserver/common/errors"
	"returntypes-langserver/common/packagetree"
)

type CodeFile struct {
	XMLName     xml.Name                 `xml:"file"`
	FilePath    string                   `xml:"path,attr"`
	PackageName string                   `xml:"package"`
	Imports     []Import                 `xml:"imports>import"`
	Classes     []*Class                 `xml:"classes>class"`
	subscriber  []packagetree.Subscriber `xml:"-"`
	parentNode  packagetree.Node         `xml:"-"`
}

// Finds a class node inside the code file.
func (codeFile *CodeFile) FindChild(name string, options packagetree.SelectionOptions) packagetree.Node {
	for i := range codeFile.Classes {
		if target := codeFile.Classes[i].TargetNode(name, options); target != nil {
			return target
		}
	}
	return nil
}

// Adds a class node to the code file.
func (codeFile *CodeFile) AddChild(node packagetree.Node, options packagetree.SelectionOptions) errors.Error {
	switch t := node.(type) {
	case *Class:
		codeFile.Classes = append(codeFile.Classes, t)
		t.parentElement = codeFile
		packagetree.ForSubscribableParents(codeFile, func(subscribable packagetree.Subscribable) {
			subscribable.OnChildAdded(t)
		})
		return nil
	}
	return errors.New(JavaErrorTitle, "Only classes supported as children of code file nodes")
}

// Removes a class node from the code file.
func (codeFile *CodeFile) RemoveChild(name string, options packagetree.SelectionOptions) errors.Error {
	for i, class := range codeFile.Classes {
		if class.TargetNode(name, options) != nil {
			codeFile.Classes[i].parentElement = nil
			if i+1 < len(codeFile.Classes) {
				// if not last element, copy all following elements to one field before
				copy(codeFile.Classes[i:], codeFile.Classes[i+1:])
			}
			// shorten slice by 1 element
			codeFile.Classes = codeFile.Classes[:len(codeFile.Classes)]
			return nil
		}
	}
	return errors.New(JavaErrorTitle, "No element to remove")
}

// Same as FindChild for the code file.
func (codeFile *CodeFile) TargetNode(name string, options packagetree.SelectionOptions) packagetree.Node {
	if name == codeFile.NodeName() {
		return codeFile
	}
	// search in the classes for a node with the name
	return codeFile.FindChild(name, options)
}

// Sets the parent of this node.
func (codeFile *CodeFile) SetParentNode(parent packagetree.Node) {
	codeFile.parentNode = parent
}

// Returns the parent of this node.
func (codeFile *CodeFile) ParentNode() packagetree.Node {
	return codeFile.parentNode
}

// Informs subscriber that a child node was added.
func (codeFile *CodeFile) OnRemove() {
	for _, subscriber := range codeFile.subscriber {
		if subscriber != nil {
			subscriber.OnRemove(codeFile)
		}
	}
	codeFile.subscriber = nil
}

// Informs its subscriber if a child node was added.
func (codeFile *CodeFile) OnChildAdded(node packagetree.Node) {
	for _, subscriber := range codeFile.subscriber {
		if subscriber != nil {
			subscriber.OnChildAdded(node, codeFile)
		}
	}
}

// Function so subscriber can register to get event information.
func (codeFile *CodeFile) Subscribe(subscriber packagetree.Subscriber) {
	for i := range codeFile.subscriber {
		if codeFile.subscriber[i] == nil {
			codeFile.subscriber[i] = subscriber
			return
		}
	}
	codeFile.subscriber = append(codeFile.subscriber, subscriber)
}

// Function so subscriber can unregister from getting event information.
func (codeFile *CodeFile) Unsubscribe(subscriber packagetree.Subscriber) {
	for i := range codeFile.subscriber {
		if codeFile.subscriber[i] == subscriber {
			codeFile.subscriber[i] = nil
			return
		}
	}
}

// Returns all methods inside the code file.
func (codeFile *CodeFile) GetAllMethods() []*Method {
	methods := make([]*Method, 0)
	for i := range codeFile.Classes {
		if codeFile.Classes[i] == nil {
			continue
		}

		subMethods := codeFile.Classes[i].GetAllMethodsInSameFile()
		if len(subMethods) > 0 {
			methods = append(methods, subMethods...)
		}
	}
	return methods
}

func (codeFile *CodeFile) Path() string {
	return codeFile.PackageName
}

// As the file name of the .java file is not directly used inside a package path, files
// are invisible nodes inside the tree. To specifically select them, the node name may be used.
func (codeFile *CodeFile) NodeName() string {
	return fmt.Sprintf("<FILE:%s>", codeFile.FilePath)
}

func (codeFile *CodeFile) Parent() JavaElement {
	return nil
}

// Returns the package tree path of a file node inside a package
func GetPackageTreePathToCodeFileNode(packageName, filePath string) string {
	return fmt.Sprintf("%s.<FILE:%s>", packageName, filePath)
}
