package projects

import (
	"path/filepath"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/processing/git"
)

type Project struct {
	configuration.Project
}

func GetProjects() []Project {
	return MapProjects(configuration.Projects())
}

func MapProjects(configuredProjects []configuration.Project) []Project {
	projects := make([]Project, 0, len(configuredProjects))
	for _, project := range configuredProjects {
		projects = append(projects, Project{
			Project: project,
		})
	}
	return projects
}

func MapConfigurationProjects(projects []Project) []configuration.Project {
	configuredProjects := make([]configuration.Project, 0, len(projects))
	for _, project := range projects {
		configuredProjects = append(configuredProjects, project.Project)
	}
	return configuredProjects
}

func (p Project) ToRepositoryDefinition() git.RepositoryDefinition {
	return git.RepositoryDefinition{
		Url:     p.GitUri,
		DirName: p.Name(),
	}
}

// Returns the path to the directory which is expected as the project's directory.
func (p Project) ExpectedDirectoryPath() string {
	if p.Directory != "" {
		return configuration.AbsolutePathFromGoProjectDir(p.Directory)
	}
	return filepath.Join(configuration.ClonerOutputDir(), p.Name())
}

func (p Project) Name() string {
	if p.AlternativeName != "" {
		return p.AlternativeName
	}
	_, repository := git.GetOwnerAndRepositoryFromURL(p.GitUri)
	return repository
}
