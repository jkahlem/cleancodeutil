package packagetree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectPackage(t *testing.T) {
	// given
	tree := CreateTestTree()

	// when
	selector := tree.Select("com.example.foo")

	// then
	assert.Equal(t, "com.example.foo", selector.Path())
	assert.Equal(t, "com.example", selector.Parent().Path())
}

func TestSelectWithHiddenNodesNode(t *testing.T) {
	// given
	tree := CreateTestTree()

	// when
	selector := tree.Select("<ROOT>.com.example.<FILE:src/main/java/Test.java>")

	// then
	assert.Equal(t, "com.example.<FILE:src/main/java/Test.java>", selector.Path())
	assert.Equal(t, "com.example", selector.Parent().Path())
}

func TestSelectorNodeCreationAtRootLevel(t *testing.T) {
	// given
	tree := CreateTestTree()

	// when
	selector := tree.Select("test")
	selector.Create()

	// then
	assert.Nil(t, selector.Err())
	assert.NotNil(t, tree.Root.FindChild("test", SelectionOptions{}))
}

func TestSelectorNodeCreationAtNestedLevel(t *testing.T) {
	// given
	tree := CreateTestTree()
	newNode := &TreeNode{name: "child"}

	// when
	selector := tree.Select("com.example.foo.bar.test")
	selector.Add(newNode)
	selector.Select("com.example.foo.bar.test.child")

	// then
	assert.Equal(t, newNode, selector.Get())
}

func TestSelectorNodeRemoving(t *testing.T) {
	// given
	tree := CreateTestTree()
	selector := tree.Select("com.example.foo")

	// when
	selector.Remove()
	newSelector := tree.Select("com.example.foo")

	// then
	assert.False(t, selector.Exists())
	assert.False(t, newSelector.Exists())
}
