package processing

import (
	"path/filepath"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/processing/git"
)

type Project struct {
	configuration.Project
}

func GetProjects() []Project {
	configuredProjects := configuration.Projects()
	projects := make([]Project, 0, len(configuredProjects))
	for _, project := range configuredProjects {
		projects = append(projects, Project{
			Project: project,
		})
	}
	return projects
}

func (p Project) ToRepositoryDefinition() git.RepositoryDefinition {
	return git.RepositoryDefinition{
		Url:     p.GitUri,
		DirName: p.ProjectName(),
	}
}

// Returns the path to the directory which is expected as the project's directory.
func (p Project) ExpectedDirectoryPath() string {
	if p.Directory != "" {
		return configuration.AbsolutePathFromGoProjectDir(p.Directory)
	}
	return filepath.Join(configuration.ClonerOutputDir(), p.ProjectName())
}

func (p Project) ProjectName() string {
	if p.AlternativeName != "" {
		return p.AlternativeName
	}
	_, repository := git.GetOwnerAndRepositoryFromURL(p.GitUri)
	return repository
}
