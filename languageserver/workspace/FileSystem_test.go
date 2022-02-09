package workspace

import (
	"testing"

	"returntypes-langserver/common/code/java"
	"returntypes-langserver/common/code/packagetree"

	"github.com/stretchr/testify/assert"
)

const SomePackage string = "com.example.somepackage"

func TestAddFile(t *testing.T) {
	// given
	filePath := "C:\\path\\to\\someFile.java"
	fileToAdd := createFile(SomePackage, filePath)
	tree := packagetree.New()
	container := NewFileSystem("C:\\", tree)

	// when
	err := container.AddFile(fileToAdd)
	codeFileNodeSelector := tree.Select(java.GetPackageTreePathToCodeFileNode(SomePackage, filePath))
	createdFile := container.GetFile(filePath)

	// then
	assert.NoError(t, err)
	assert.True(t, codeFileNodeSelector.Exists())
	assert.NotNil(t, createdFile)
}

func TestRenameFile(t *testing.T) {
	// given
	oldPath := "C:\\path\\to\\someFile.java"
	newPath := "C:\\path\\to\\someRenamedFile.java"
	container := createFileContainer(createFile(SomePackage, oldPath))

	// when
	err := container.RenameFile(oldPath, newPath)
	fileAtOldLocation := container.GetFile(oldPath)
	fileAtNewLocation := container.GetFile(newPath)

	// then
	assert.NoError(t, err)
	assert.Nil(t, fileAtOldLocation)
	assert.NotNil(t, fileAtNewLocation)
}

func TestRemoveFile(t *testing.T) {
	// given
	filePath := "C:\\path\\to\\someFile.java"
	container := createFileContainer(createFile(SomePackage, filePath))

	// when
	err := container.RemoveFile(filePath)
	file := container.GetFile(filePath)
	codeFileNodeSelector := container.tree.Select(java.GetPackageTreePathToCodeFileNode(SomePackage, filePath))

	// then
	assert.NoError(t, err)
	assert.Nil(t, file)
	assert.False(t, codeFileNodeSelector.Exists())
}

func TestReplaceFile(t *testing.T) {
	// given
	filePath := "C:\\path\\to\\someFile.java"
	differentPackage := "com.example.different.package"
	container := createFileContainer(createFile(SomePackage, filePath))

	// when
	err := container.ReplaceFile(createFile(differentPackage, filePath))
	file := container.GetFile(filePath)

	// then
	assert.NoError(t, err)
	assert.NotNil(t, file)
	assert.Equal(t, file.file.PackageName, differentPackage)
}

// Helper functions

func createFileContainer(fileToAdd FileWrapper) VirtualFileSystem {
	tree := packagetree.New()
	container := NewFileSystem("C:\\", tree)
	container.AddFile(fileToAdd)
	return container
}

func createFile(packageName, filePath string) FileWrapper {
	wrapper := FileWrapper{}
	wrapper.file = &java.CodeFile{
		PackageName: packageName,
		FilePath:    filePath,
	}
	return wrapper
}
