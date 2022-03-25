package packagetree

import (
	"regexp"
	"strings"

	"returntypes-langserver/common/debug/errors"
)

const PackageTreeErrorTitle = "Error"

// Selection parts are splitted by dots.
// Selector nodes are selectable by using special selectors with <> (Content depends on node implemention)
const SelectionPartPattern = "(<.+?>|.+?)(\\.|$)"

// The selection options define further options for the operations on the selector.
type SelectionOptions struct {
	// If true, the nodes may create new nodes on the path if the next node was not found.
	// But: This does NOT ensure that the returning pointer will be other than nil.
	// There may be nodes which can not create subnodes if they do not exist (e.g. classes)
	CreateEmptyNodeIfNotExists bool
	// If true, the search will not go through private classes
	FindOnlyPublicClasses bool
	// If true, fires no events to subscriber (makes building the tree on the first time faster)
	Silent bool
}

// Represents a part of the package path.
type SelectionPart struct {
	name string
	node Node
}

// The Selector object makes it easier to work on the Package Tree.
// It is possible to select certain nodes, add new ones and remove existing ones or replace some nodes.
//
// If an error occurs, the methods of the selector will have no effects on the package tree anymore,
// and the error can be checked using the Err() function.
type Selector struct {
	selection []SelectionPart
	options   SelectionOptions
	tree      *Tree
	err       errors.Error
}

// Creates a new selector.
func NewSelector(options SelectionOptions) Selector {
	return Selector{
		options: options,
	}
}

// Sets the selector on a tree.
func (selector *Selector) SetTree(tree *Tree) {
	selector.tree = tree

	// when changing the tree, the nodes on the path may also change, so remove them from the selection
	for i := range selector.selection {
		selector.selection[i].node = nil
	}
}

// Selects the specified path for working.
func (selector *Selector) Select(path string) {
	if selector.Err() != nil {
		return
	}

	re := regexp.MustCompile(SelectionPartPattern)
	splittedPath := re.FindAllString(path, -1)
	if splittedPath[0] == "<ROOT>." {
		splittedPath = splittedPath[1:]
	}
	selector.selection = make([]SelectionPart, len(splittedPath))
	for i := range selector.selection {
		strlen := len(splittedPath[i])
		if strlen > 0 && splittedPath[i][strlen-1] == '.' {
			// remove the trailing dot
			splittedPath[i] = splittedPath[i][:strlen-1]
		}
		selector.selection[i].name = splittedPath[i]
	}
}

// Returns a selector for the parent path.
func (selector *Selector) Parent() *Selector {
	if len(selector.selection) == 0 {
		return nil
	}

	// create a copy for parent
	var parentSelector Selector = *selector
	parentElementCount := len(selector.selection) - 1
	parentSelector.selection = make([]SelectionPart, parentElementCount)
	copy(parentSelector.selection, selector.selection[:len(selector.selection)])
	return &parentSelector
}

// The selector tries to create a package tree node at the specified position. For more specific nodes, use Add().
func (selector *Selector) Create() {
	if selector.Err() != nil || !selector.checkValidSelection() {
		return
	}

	previousSetting := selector.options.CreateEmptyNodeIfNotExists
	selector.options.CreateEmptyNodeIfNotExists = true
	node := selector.Get()
	selector.options.CreateEmptyNodeIfNotExists = previousSetting

	if node == nil && selector.Err() != nil {
		selector.err = errors.New(PackageTreeErrorTitle, "Could not create a new node")
	}
}

// Adds the node at the selected path.
func (selector *Selector) Add(node Node) {
	if selector.Err() != nil || !selector.checkTree() || !selector.checkValidSelection() {
		return
	}

	//parentSelector := selector.Parent()
	previousSetting := selector.options.CreateEmptyNodeIfNotExists
	selector.options.CreateEmptyNodeIfNotExists = true
	parentNode := selector.Get()
	selector.options.CreateEmptyNodeIfNotExists = previousSetting
	if selector.Err() != nil {
		return
	}

	if parentNode == nil {
		selector.err = errors.New(PackageTreeErrorTitle, "Could not find parent node at "+selector.Path())
		return
	}

	parentNode.AddChild(node, selector.options)
}

// Removes the node at the selected path.
func (selector *Selector) Remove() {
	if selector.Err() != nil || !selector.checkTree() || !selector.checkValidSelection() {
		return
	}

	parentNode := selector.Parent().Get()

	if parentNode == nil {
		selector.err = errors.New(PackageTreeErrorTitle, "Could not find parent node at "+selector.Path())
		return
	}

	parentNode.RemoveChild(selector.Name(), selector.options)
	// remove the node of the last selection part (if it was cached before by Get())
	selector.selection[len(selector.selection)-1].node = nil
}

// Gets the currently selected node. May be nil if no node was found.
func (selector *Selector) Get() Node {
	if selector.Err() != nil {
		// return last accepted node or nil
		return nil
	}

	if selector.tree == nil {
		selector.err = errors.New(PackageTreeErrorTitle, "No tree root set")
		return nil
	}

	// walk the selection path and search for the nodes
	var current Node = selector.tree.Root
	for i, selection := range selector.selection {
		if selection.node != nil {
			// if the node was already found (for example, by a previous Get() call) use it
			current = selection.node
		} else {
			child := current.FindChild(selection.name, selector.options)
			if child != nil {
				current = child
			} else if selector.options.CreateEmptyNodeIfNotExists {
				// maybe create a new package tree node if a node does not exist
				child = &TreeNode{
					name: selection.name,
				}
				current.AddChild(child, selector.options)
				current = child
			} else {
				selector.err = errors.New(PackageTreeErrorTitle, "Selector Error: Node %s not found in path %s", selection.name, selector.Path())
				return nil
			}

			selector.selection[i].node = current
		}
	}
	return current
}

// Checks if a node exists at the current path.
func (selector *Selector) Exists() bool {
	return selector.Get() != nil
}

// The currently selected path as a string seperated with dots.
func (selector *Selector) Path() string {
	path := make([]string, len(selector.selection))
	for i, part := range selector.selection {
		path[i] = part.name
	}
	return strings.Join(path, ".")
}

// The name of the targeted node (the last part of the package path).
func (selector *Selector) Name() string {
	if len(selector.selection) == 0 {
		return ""
	}
	return selector.selection[len(selector.selection)-1].name
}

// Returns the first error which stopped the selector from working..
func (selector *Selector) Err() errors.Error {
	return selector.err
}

// Checks if the tree is set and returns true, otherwise sets an error and returns false.
func (selector *Selector) checkTree() bool {
	if selector.tree == nil {
		selector.err = errors.New(PackageTreeErrorTitle, "No tree set")
		return false
	}
	return true
}

// Checks whether the selection is valid and returns true, otherwise sets an error and returns false.
func (selector *Selector) checkValidSelection() bool {
	if len(selector.selection) == 0 {
		selector.err = errors.New(PackageTreeErrorTitle, "No selection")
		return false
	}
	return true
}
