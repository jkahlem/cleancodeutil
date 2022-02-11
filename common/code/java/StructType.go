package java

import (
	"returntypes-langserver/common/code/packagetree"
	"strings"
)

// The type object represents the type of a java element like a method or extended classes.
// It will point to the definition of a type in the package tree or a primitive type. If
// the type definition is removed from the package tree, the resolution state of the type
// will be resetted and needs to be resolved again (if possible).
type Type struct {
	TypeName            string      `xml:",chardata"`
	IsArrayType         bool        `xml:"isArrayType,attr"`
	parentElement       JavaElement `xml:"-"`
	resolutionTypeName  string
	TypeResolutionState ResolutionState
}

// If the node this type was pointing to was removed, reset the resolution state of this type.
func (javaType *Type) OnRemove(node packagetree.Node) {
	if root := packagetree.FindRoot(node); root != nil {
		if treeNode, ok := root.(*packagetree.TreeNode); ok {
			treeNode.AddSubscriber(javaType)
		}
	}
	// Type need to be resolved again after removal.
	javaType.TypeResolutionState = Unresolved
	javaType.resolutionTypeName = javaType.TypeName
}

func (javaType *Type) OnChildAdded(node, subscribable packagetree.Node) {}

func (javaType *Type) Path() string {
	if javaType.Parent() == nil {
		return javaType.TypeName
	}
	return strings.Join([]string{javaType.Parent().Path(), javaType.TypeName}, ".")
}

func (javaType *Type) Parent() JavaElement {
	return javaType.parentElement
}
