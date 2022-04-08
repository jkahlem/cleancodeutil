package extractor

import (
	"io"
	"os"
	"path/filepath"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/common/utils"
	"returntypes-langserver/processing/projects"
	"returntypes-langserver/services/crawler"
)

// Preprocesses the java code for one project
func PreprocessSourceCodeForProject(project projects.Project) {
	// If an output file does already exist, skip preprocessing the data for this project.
	if exists, err := preprocessedSourceCodeFileExists(project); err != nil {
		log.ReportProblemWithError(err, "Could not check if xml output file for %s exists", project.ProjectName())
		return
	} else if exists {
		return
	}

	// Use the crawler to preprocess the java code structures for a given project into one xml file
	log.Info("Preprocess java code for project %s\n", project.ProjectName())
	if !utils.DirExists(project.ExpectedDirectoryPath()) {
		log.ReportProblem("Skip project %s as it does not exist at %s\n", project.ProjectName(), project.ExpectedDirectoryPath())
		return
	}

	if xml, err := crawlProject(project); err != nil {
		log.ReportProblemWithError(err, "Could not create output file for java code files")
		return
	} else if err := savePreprocessedXmlContent(project, xml); err != nil {
		log.ReportProblemWithError(err, "Could not write to output file for java code files")
		return
	}
}

func crawlProject(project projects.Project) (string, errors.Error) {
	javaVersion := project.JavaVersion
	if javaVersion == 0 {
		javaVersion = configuration.CrawlerDefaultJavaVersion()
	}
	crawlerOptions := crawler.NewOptions().
		Forced(!configuration.StrictMode()).
		WithJavaVersion(javaVersion).
		Build()
	return crawler.GetRawCodeElementsOfDirectory(project.ExpectedDirectoryPath(), crawlerOptions)
}

func savePreprocessedXmlContent(project projects.Project, xml string) errors.Error {
	// Write the preprocessed code structures to an xml file
	file, err := os.Create(GetPreprocessedFilePathForProject(project))
	defer file.Close()
	if err != nil {
		return errors.Wrap(err, "Error", "Could not create output file")
	} else if _, err := io.WriteString(file, xml); err != nil {
		return errors.Wrap(err, "Error", "Could not write tooutput file")
	}
	return nil
}

// Returns true if the crawler output file for the given project does exist
func preprocessedSourceCodeFileExists(project projects.Project) (bool, errors.Error) {
	_, err := os.Stat(GetPreprocessedFilePathForProject(project))
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, errors.Wrap(err, "Error", "Unexpected file error")
}

func GetPreprocessedFilePathForProjects(projects []projects.Project) []string {
	directories := make([]string, 0, len(projects))
	for _, project := range projects {
		directories = append(directories, GetPreprocessedFilePathForProject(project))
	}
	return directories
}

func GetPreprocessedFilePathForProject(project projects.Project) string {
	return filepath.Join(configuration.CrawlerOutputDir(), project.ProjectName()+".xml")
}
