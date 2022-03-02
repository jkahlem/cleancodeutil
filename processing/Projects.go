package processing

import (
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
	_, repository := git.GetOwnerAndRepositoryFromURL(p.GitUri)
	return git.RepositoryDefinition{
		Url:     p.GitUri,
		DirName: repository,
	}
}
