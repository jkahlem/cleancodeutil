// This package allows managing packages/classes etc. inside an undirected acyclic graph.
// It is here stated as a tree (as packages/classes are usually ordered as a tree)
// but the methods of the Tree object will not ensure the validity of the tree to save
// resources.
//
// However, interacting with the tree using the Selector class is a safe way
// to ensure the tree structure while having an easier to use high level interface for
// manipulating the tree. Therefore, tree manipulation should only happen using selectors.
package packagetree

import "returntypes-langserver/common/debug/errors"

type Tree struct {
	Root *TreeNode
}

type TreeNode struct {
	subNodes   []Node
	name       string
	parent     Node
	subscriber []Subscriber
}

// Represents a node
type Node interface {
	// searches for child nodes with the specified name
	FindChild(name string, options SelectionOptions) Node
	// adds a node as children. If this is not possible, the function should return an error
	AddChild(node Node, options SelectionOptions) errors.Error
	// removes a child from the node. If this is not possible (for example: the node does not exist anymore) it should report an error
	RemoveChild(name string, options SelectionOptions) errors.Error
	// returns the node represented by the name.. this is especially for invisible nodes like java.CodeFile
	TargetNode(name string, options SelectionOptions) Node
	// gets the parent of this node
	ParentNode() Node
	// sets the parent of this node
	SetParentNode(Node)
	// the (selection) path to the node
	NodeName() string
}

type Subscribable interface {
	// informs its subscriber that this node was detached
	OnRemove()
	// informs its subscriber if a child node was added
	OnChildAdded(Node)
	// function so subscriber can register to get event information
	AddSubscriber(Subscriber)
	// function for subscriber to unsubscribe
	RemoveSubscriber(Subscriber)
}

type Subscriber interface {
	// if the subscribed node was detach
	OnRemove(subscribable Node)
	// if a node was added to the subscribed node
	OnChildAdded(child, subscribable Node)
}

// Creates a new empty package tree.
func New() Tree {
	return Tree{Root: NewNode("")}
}

// Creates a new empty package tree node with the specified name.
func NewNode(name string) *TreeNode {
	return &TreeNode{
		name: name,
	}
}

// Searches for a root node.
func FindRoot(node Node) Node {
	if node == nil {
		return nil
	}

	for node.ParentNode() != nil {
		node = node.ParentNode()
	}
	return node
}

// Creates a full path inside the tree until the given node. The returned path will also contain all invisible nodes.
func PathToNode(node Node) string {
	path := node.NodeName()
	for node.ParentNode() != nil {
		parent := node.ParentNode()
		path = parent.NodeName() + "." + path
		node = parent
	}
	return path
}

// Execute the function for all subscribable parent nodes of node (including the node itself).
func ForSubscribableParents(node Node, fn func(Subscribable)) {
	for node != nil {
		if subscribable, ok := node.(Subscribable); ok {
			fn(subscribable)
			//subscribable.OnChildAdded(child)
		}
		node = node.ParentNode()
	}
}

// Returns true if the node is a root node.
func (node *TreeNode) IsRoot() bool {
	return node.parent == nil
}

// The path of the node inside the tree.
func (node *TreeNode) NodeName() string {
	if node.IsRoot() {
		return "<ROOT>"
	}
	return node.name
}

// Sets the parent of this node.
func (node *TreeNode) SetParentNode(parent Node) {
	node.parent = parent
}

// Returns the parent of this node.
func (node *TreeNode) ParentNode() Node {
	return node.parent
}

// Returns this node if it matches the name.
func (node *TreeNode) TargetNode(name string, options SelectionOptions) Node {
	if node.name == name {
		return node
	}
	return nil
}

// Searches a sub node with the given name.
func (node *TreeNode) FindChild(name string, options SelectionOptions) Node {
	for _, child := range node.subNodes {
		if child != nil {
			if target := child.TargetNode(name, options); target != nil {
				return target
			}
		}
	}
	return nil
}

// Adds the node to this node.
func (node *TreeNode) AddChild(child Node, options SelectionOptions) errors.Error {
	defer node.afterAddChild(child, options)

	// search for a free place in the slice
	for i, existingChild := range node.subNodes {
		if existingChild == nil {
			node.subNodes[i] = child
			return nil
		}
	}

	// otherwise append it to the last
	node.subNodes = append(node.subNodes, child)
	return nil
}

// Sets parents and informs subscribable parent nodes about the new child node.
func (node *TreeNode) afterAddChild(child Node, options SelectionOptions) {
	child.SetParentNode(node)
	if !options.Silent {
		ForSubscribableParents(node, func(subscribable Subscribable) {
			subscribable.OnChildAdded(child)
		})
	}
}

// Removes the node with the given name from this node's children.
func (node *TreeNode) RemoveChild(name string, options SelectionOptions) errors.Error {
	for i, child := range node.subNodes {
		if target := child.TargetNode(name, options); target != nil {
			if target == child {
				node.subNodes[i] = nil
				child.SetParentNode(nil)

				// inform child that it was detached
				if subscribable, ok := child.(Subscribable); ok {
					subscribable.OnRemove()
				}
			} else {
				child.RemoveChild(name, options)
			}
		}
	}
	return nil
}

// Informs it's subscriber that this node was detached.
func (node *TreeNode) OnRemove() {
	for _, subscriber := range node.subscriber {
		if subscriber != nil {
			subscriber.OnRemove(node)
		}
	}
	node.subscriber = nil
}

// Informs it's subscriber if a child node was added.
func (node *TreeNode) OnChildAdded(child Node) {
	for _, subscriber := range node.subscriber {
		if subscriber != nil {
			subscriber.OnChildAdded(child, node)
		}
	}
}

// Registers a subscriber for event information.
func (node *TreeNode) AddSubscriber(subscriber Subscriber) {
	for i := range node.subscriber {
		if node.subscriber[i] == nil {
			node.subscriber[i] = subscriber
			return
		}
	}
	node.subscriber = append(node.subscriber, subscriber)
}

// Unregisters a subscriber for event information.
func (node *TreeNode) RemoveSubscriber(subscriber Subscriber) {
	for i := range node.subscriber {
		if node.subscriber[i] == subscriber {
			node.subscriber[i] = nil
		}
	}
}

// Shorthand for selecting a specific node on the tree using a path with default options.
func (t *Tree) Select(path string) Selector {
	return t.SelectWithOptions(path, SelectionOptions{})
}

// Shorthand for selecting a specific node on the tree with default options.
func (t *Tree) SelectNode(node Node) Selector {
	return t.SelectWithOptions(PathToNode(node), SelectionOptions{})
}

// Shorthand for selecting a specific path on the tree with custom selection options.
func (t *Tree) SelectWithOptions(path string, options SelectionOptions) Selector {
	selector := NewSelector(options)
	selector.SetTree(t)
	selector.Select(path)
	return selector
}
