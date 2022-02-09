package git

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/debug/log"
	"strings"

	"returntypes-langserver/common/debug/errors"
)

// Loads the repositories to clone from the git input file and clones them to the project input dir (if not already exist)
func CloneRepositories() errors.Error {
	if err := createRepositoryOutputDir(); err != nil {
		return err
	}

	list := GetRepositoryList()
	for _, repository := range list.Repositories {
		if cloned, err := isAlreadyClonedRepository(repository.DirName); err != nil {
			log.ReportProblemWithError(err, "Skipped cloning %s because an error occured", repository.DirName)
			continue
		} else if cloned {
			continue
		}

		log.Info("Clone repository at %s as %s\n", repository.Url, repository.DirName)
		if err := CloneRepository(repository); err != nil {
			log.ReportProblemWithError(err, "Skipped cloning %s because an error occured", repository.DirName)
			continue
		}
	}
	return nil
}

// Parses the repository url and the name of the output directory from a line of the repository list
func parseRepositoryListLine(line string) (url, name string) {
	parts := strings.Split(line, " ")
	if len(parts) == 1 {
		url = parts[0]
		_, name = getOwnerAndRepositoryFromURL(url)
	} else if len(parts) > 1 {
		url = parts[0]
		name = parts[1]
	}
	return
}

// Creates the repository output dir if it does not already exist
func createRepositoryOutputDir() errors.Error {
	if _, err := os.Stat(configuration.ProjectInputDir()); os.IsNotExist(err) {
		if err = os.MkdirAll(configuration.ProjectInputDir(), os.ModePerm); err != nil {
			return errors.Wrap(err, CloneErrorTitle, "Could not create output directory")
		}
	} else if err != nil {
		return errors.Wrap(err, CloneErrorTitle, "Unexpected file error")
	}
	return nil
}

// checks if the repository is already cloned (inside the project input dir)
func isAlreadyClonedRepository(repositoryName string) (bool, errors.Error) {
	fileInfo, err := os.Stat(filepath.Join(configuration.ProjectInputDir(), repositoryName))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, errors.Wrap(err, CloneErrorTitle, "Unexpected file error")
	}
	return fileInfo.IsDir(), nil
}

func getOwnerAndRepositoryFromURL(url string) (owner, repository string) {
	splitted := strings.Split(url, "/")
	repository = splitted[len(splitted)-1]
	owner = splitted[len(splitted)-2]
	return
}

func LoadRepositoryInfo(repositoryName string) ([]byte, errors.Error) {
	path := filepath.Join(getPathToRepositoryCloneDir(repositoryName), "repositoryInfo.json")
	if content, err := ioutil.ReadFile(path); err != nil {
		return content, errors.Wrap(err, CloneErrorTitle, "Could not read repository info file")
	} else {
		return content, nil
	}
}
