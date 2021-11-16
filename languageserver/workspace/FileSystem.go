package workspace

import (
	"returntypes-langserver/common/errors"
	"returntypes-langserver/common/java"
	"returntypes-langserver/common/packagetree"
)

// Manages file operations on the virtual file system of a workspace.
type VirtualFileSystem struct {
	root  string
	files map[string]*FileWrapper
	cache WorkspaceFileCache
	tree  packagetree.Tree
}

// Caches the files in the virtual file system for list representations.
type WorkspaceFileCache struct {
	CodeFiles  []*java.CodeFile
	Files      []*FileWrapper
	IsUpToDate bool
}

// Creates a new virtual file system
func NewFileSystem(rootPath string, tree packagetree.Tree) VirtualFileSystem {
	filesys := VirtualFileSystem{
		root: rootPath,
		tree: tree,
	}
	filesys.prepare()
	return filesys
}

func (filesys *VirtualFileSystem) prepare() {
	filesys.files = make(map[string]*FileWrapper)
}

// Adds a file to the virtual file system
func (filesys *VirtualFileSystem) AddFile(wrapper FileWrapper) errors.Error {
	if wrapper.Exists() {
		// create entry and set cache out of date
		filesys.cache.IsUpToDate = false
		filesys.files[wrapper.Path()] = &wrapper

		// create entry in package tree
		if len(wrapper.file.PackageName) > 0 {
			selector := filesys.tree.Select(wrapper.file.PackageName)
			selector.Add(wrapper.file)
			if selector.Err() != nil {
				return selector.Err()
			}
		}
	}
	return nil
}

// Renames a file in the virtual file system
func (filesys *VirtualFileSystem) RenameFile(oldPath, newPath string) errors.Error {
	if file := filesys.GetFile(oldPath); file != nil {
		// create entry and set cache out of date
		filesys.cache.IsUpToDate = false
		delete(filesys.files, file.Path())
		filesys.files[newPath] = file
	}
	return nil
}

// Returns the file with the given path inside the virtual file system.
func (filesys *VirtualFileSystem) GetFile(path string) *FileWrapper {
	return filesys.files[path]
}

// Removes a file of the virtual file system.
func (filesys *VirtualFileSystem) RemoveFile(path string) errors.Error {
	existingFile := filesys.GetFile(path)
	if existingFile == nil {
		return nil
	}

	selector := filesys.tree.SelectNode(existingFile.file)
	selector.Remove()
	if selector.Err() != nil {
		return selector.Err()
	}

	filesys.files[path] = nil
	filesys.cache.IsUpToDate = false
	return nil
}

// Replaces a file in the virtual file system.
func (filesys *VirtualFileSystem) ReplaceFile(wrapper FileWrapper) errors.Error {
	if !wrapper.Exists() {
		return errors.New(WorkspaceErrorTitle, "The given file does not exist")
	}

	if err := filesys.RemoveFile(wrapper.file.FilePath); err != nil {
		return err
	}
	err := filesys.AddFile(wrapper)
	return err
}

// Returns all java files in the virtual file system
func (filesys *VirtualFileSystem) CodeFiles() []*java.CodeFile {
	filesys.updateCache()
	return filesys.cache.CodeFiles
}

// Returns all files in the virtual file system
func (filesys *VirtualFileSystem) Files() []*FileWrapper {
	filesys.updateCache()
	return filesys.cache.Files
}

func (filesys *VirtualFileSystem) PackageTree() *packagetree.Tree {
	return &filesys.tree
}

func (filesys *VirtualFileSystem) updateCache() {
	if filesys.cache.IsUpToDate {
		return
	}

	codeFiles := make([]*java.CodeFile, len(filesys.files))
	wrappers := make([]*FileWrapper, len(filesys.files))
	i := 0
	for _, wrapper := range filesys.files {
		if wrapper == nil || wrapper.file == nil {
			continue
		}

		codeFiles[i] = wrapper.file
		wrappers[i] = wrapper
		i++
	}
	filesys.cache.CodeFiles = codeFiles[:i]
	filesys.cache.Files = wrappers[:i]
	filesys.cache.IsUpToDate = true
}
