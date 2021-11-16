package workspace

// A container for workspaces
type Container struct {
	workspaces []*Workspace
}

// Finds a workspace with the given root path
func (w *Container) FindWorkspace(path string) *Workspace {
	for i, workspace := range w.workspaces {
		if workspace.RootPath() == path {
			return w.workspaces[i]
		}
	}
	return nil
}

// Returns true if a workspace with the given root path exists
func (w *Container) WorkspaceExists(path string) bool {
	return w.FindWorkspace(path) != nil
}

// Creates a workspace on the given root path
func (w *Container) CreateWorkspace(path string) *Workspace {
	if w.WorkspaceExists(path) {
		return nil
	}

	workspace := New(path)
	w.workspaces = append(w.workspaces, &workspace)
	return &workspace
}

// Returns a list of workspaces in the container
func (w *Container) List() []*Workspace {
	slice := make([]*Workspace, len(w.workspaces))
	copy(slice, w.workspaces)
	return slice
}
