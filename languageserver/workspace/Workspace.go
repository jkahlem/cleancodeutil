package workspace

import (
	"strings"

	"returntypes-langserver/common/code/java"
	"returntypes-langserver/common/code/packagetree"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/services/crawler"
)

const WorkspaceErrorTitle = "Workspace Error"

type Workspace struct {
	FileSystem VirtualFileSystem
}

// Creates a new workspace
func New(rootPath string) Workspace {
	ws := Workspace{}
	ws.createFileSystem(rootPath)
	return ws
}

func (w *Workspace) RootPath() string {
	return w.FileSystem.root
}

func (w *Workspace) createFileSystem(rootPath string) {
	tree := w.createPackageTree(rootPath)
	w.FileSystem = NewFileSystem(rootPath, tree)
}

func (w *Workspace) createPackageTree(rootPath string) packagetree.Tree {
	tree := packagetree.New()
	java.LoadDefaultPackagesToTree(&tree)
	return tree
}

// Loads the files in the workspace.
func (w *Workspace) Load() errors.Error {
	if err := w.LoadFilesInsideWorkspace(); err != nil {
		return err
	}
	if err := java.LoadDefaultPackagesToTree(&w.FileSystem.tree); err != nil {
		return err
	}

	return nil
}

// Loads all files of the workspace into the virtual workspace.
func (w *Workspace) LoadFilesInsideWorkspace() errors.Error {
	fileContainer, err := crawler.GetCodeElementsOfDirectory(w.RootPath(), w.crawlerOptions())
	if err != nil {
		return errors.Wrap(err, WorkspaceErrorTitle, "Loading error in workspace")
	}

	for i := range fileContainer.CodeFiles() {
		wrapper := FileWrapper{
			file: fileContainer.CodeFiles()[i],
		}
		if err := w.FileSystem.AddFile(wrapper); err != nil {
			return errors.Wrap(err, WorkspaceErrorTitle, "Loading error in workspace")
		}
	}

	return nil
}

// Adds a file to the virtual workspace.
func (w *Workspace) AddFile(path string) errors.Error {
	fileContainer, err := crawler.GetCodeElements(path, w.crawlerOptions())
	if err != nil {
		return errors.Wrap(err, WorkspaceErrorTitle, "Could not load file")
	}
	if len(fileContainer.CodeFiles()) == 0 {
		return nil
	}

	wrapper := FileWrapper{
		file: fileContainer.CodeFiles()[0],
	}
	if err := w.FileSystem.AddFile(wrapper); err != nil {
		return errors.Wrap(err, WorkspaceErrorTitle, "Loading error in workspace")
	}

	return nil
}

// Reloads a file into the virtual workspace.
func (w *Workspace) ReloadFile(path string) errors.Error {
	fileContainer, err := crawler.GetCodeElements(path, w.crawlerOptions())
	if err != nil {
		return errors.Wrap(err, WorkspaceErrorTitle, "Could not reload file")
	}
	if len(fileContainer.CodeFiles()) == 0 {
		return nil
	}

	targetFile := fileContainer.CodeFiles()[0]
	err = w.FileSystem.ReplaceFile(FileWrapper{
		file: targetFile,
	})
	return err
}

// Renames a file inside the virtual workspace.
func (w *Workspace) RenameFile(oldPath, newPath string) {
	w.FileSystem.RenameFile(oldPath, newPath)
}

// Returns true if the file should belong to the virtual workspace but this does NOT mean that the file is actually
// inside the workspace.
func (w *Workspace) IsFileBelongingToWorkspace(path string) bool {
	return strings.HasPrefix(path, w.RootPath())
}

func (w *Workspace) crawlerOptions() crawler.Options {
	return crawler.NewOptions().WithRanges(true).WithAbsolutePaths(true).Silent(true).Forced(true).Build()
}
