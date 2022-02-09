package packagetree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTreeNodeFindChild(t *testing.T) {
	// given
	tree := CreateTestTree()
	noopt := SelectionOptions{}

	// when
	child := tree.Root.FindChild("child1", noopt)

	// then
	assert.NotNil(t, child)
}

func TestTreeNodeFindNestedChild(t *testing.T) {
	// given
	tree := CreateTestTree()
	noopt := SelectionOptions{}

	// when
	child := tree.Root.FindChild("com", noopt).FindChild("example", noopt).FindChild("foo", noopt)

	// then
	assert.NotNil(t, child)
}

func TestTreeNodeRemoveNestedChild(t *testing.T) {
	// given
	tree := CreateTestTree()
	noopt := SelectionOptions{}

	// when
	tree.Root.FindChild("com", noopt).RemoveChild("example", noopt)
	child := tree.Root.FindChild("com", noopt).FindChild("example", noopt)

	// then
	assert.Nil(t, child)
}

func TestInformSubscribersOnAddChild(t *testing.T) {
	// given
	subscribable := NewNode("subscribable")
	subscriber := TestSubscriber{}
	childToAdd := NewNode("newNode")
	subscribable.AddSubscriber(&subscriber)

	// when
	subscribable.AddChild(childToAdd, SelectionOptions{})

	// then
	assert.Equal(t, childToAdd, subscriber.AddedChildNode)
}

func TestInformSubscribersOnRemoval(t *testing.T) {
	// given
	root := NewNode("subscribable")
	child := NewNode("newNode")
	root.AddChild(child, SelectionOptions{})
	subscriber := TestSubscriber{}
	child.AddSubscriber(&subscriber)

	// when
	root.RemoveChild(child.name, SelectionOptions{})

	// then
	assert.True(t, subscriber.OnRemoveCalled)
}

// Test relevant structures

type TestSubscriber struct {
	AddedChildNode Node
	OnRemoveCalled bool
}

func (s *TestSubscriber) OnRemove(subscribable Node) {
	s.OnRemoveCalled = true
}

func (s *TestSubscriber) OnChildAdded(child, subscribable Node) {
	s.AddedChildNode = child
}

// Helper functions

func CreateChildNode(name string) *TreeNode {
	return &TreeNode{
		name: name,
	}
}

func CreateTestTree() *Tree {
	// a tree with child1, child2 and com.example.foo
	tree := New()
	noopt := SelectionOptions{}
	tree.Root.AddChild(CreateChildNode("child1"), noopt)
	tree.Root.AddChild(CreateChildNode("child2"), noopt)

	com := CreateChildNode("com")
	tree.Root.AddChild(com, noopt)
	comExample := CreateChildNode("example")
	com.AddChild(comExample, noopt)
	comExampleFoo := CreateChildNode("foo")
	comExample.AddChild(comExampleFoo, noopt)

	return &tree
}
