package git

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/common/utils"
	"strings"
)

const CloneErrorTitle = "Clone Error"
const LargeRepositorySize = 1024 * 128

type Cloner interface {
	Clone(uri, outputDir string) errors.Error
}

// Contains repository info
type RepositoryInfoWrapper struct {
	// Repository info as the struct (has only needed fields)
	Info RepositoryInfo
	// The raw response data from github API
	Raw []byte
}

// The repository info
type RepositoryInfo struct {
	// The expected size of the repository in KB. The actual size might exceed this value due to the way Github API handles this value.
	Size int `json:"size"`
}

// Defines a repository to clone
type RepositoryDefinition struct {
	Url     string `json:"url"`
	DirName string `json:"dirName"`
}

// Clones the repository with the given name. The clone process may wait a bit (to prevent github from rejecting the cloning)
func CloneRepository(repository RepositoryDefinition) errors.Error {
	info, _ := getRepositoryInfo(repository)
	if !checkSize(info) {
		log.Info("Skip cloning of %s because the repository size (%s) exceeds the maximum cloning size. (%s)\n",
			repository.Url, fmtSize(info.Info.Size), fmtSize(configuration.ClonerMaximumCloneSize()))
		return nil
	} else if info != nil && info.Info.Size > LargeRepositorySize {
		log.Info("Clone process may take a while because of cloning a large repository. (Size: %s)\n", fmtSize(info.Info.Size))
	}
	if err := clone(repository); err != nil {
		return err
	} else if err := createRepositoryInfoFile(repository, info); err != nil {
		return err
	}

	return nil
}

// Returns the repository info of a repository
func getRepositoryInfo(repository RepositoryDefinition) (*RepositoryInfoWrapper, errors.Error) {
	if strings.HasPrefix(repository.Url, "https://github.com") {
		owner, repositoryName := GetOwnerAndRepositoryFromURL(repository.Url)
		if raw, err := getRepositoryInfoFromGithubAPI(owner, repositoryName); err != nil {
			return nil, err
		} else {
			var info RepositoryInfo
			if err := json.Unmarshal(raw, &info); err != nil {
				return nil, errors.Wrap(err, CloneErrorTitle, "Could not parse repository info")
			}
			return &RepositoryInfoWrapper{
				Info: info,
				Raw:  raw,
			}, nil
		}
	}
	return nil, nil
}

// Returns the repository info from a repository from github API
func getRepositoryInfoFromGithubAPI(owner, repositoryName string) ([]byte, errors.Error) {
	response, err := http.Get("https://api.github.com/repos/" + owner + "/" + repositoryName)
	if err != nil {
		return nil, errors.Wrap(err, CloneErrorTitle, "Could not get repository info")
	}

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, CloneErrorTitle, "Could not get repository info")
	}
	return content, nil
}

// Returns true if the repository size does not exceed the maximum clone size.
func checkSize(info *RepositoryInfoWrapper) bool {
	if info != nil && info.Info.Size > configuration.ClonerMaximumCloneSize() {
		return false
	}
	return true
}

// Clones the repository.
func clone(repository RepositoryDefinition) errors.Error {
	return getCloner().Clone(repository.Url, getPathToRepositoryCloneDir(repository.DirName))
}

// Returns the cloner to use for the cloning process
func getCloner() Cloner {
	if configuration.ClonerUseCommandLineTool() {
		return &CommandLineCloner{}
	}
	return &IntegratedCloner{}
}

// Returns a path to the expected repository dir in the project input dir
func getPathToRepositoryCloneDir(repositoryName string) string {
	return filepath.Join(configuration.ClonerOutputDir(), repositoryName)
}

// Creates a repository info file from the github API
func createRepositoryInfoFile(repository RepositoryDefinition, info *RepositoryInfoWrapper) errors.Error {
	if info != nil {
		if err := saveRepositoryInfo(info.Raw, repository.DirName); err != nil {
			return err
		}
	}
	return nil
}

// Saves the repository info in a json file in the repository directory
func saveRepositoryInfo(repositoryInfo []byte, repositoryName string) errors.Error {
	path := filepath.Join(getPathToRepositoryCloneDir(repositoryName), "repositoryInfo.json")
	if err := ioutil.WriteFile(path, repositoryInfo, 0777); err != nil {
		return errors.Wrap(err, CloneErrorTitle, "Could not write repository info")
	}
	return nil
}

// Returns the passed size in kilobytes formatted in the IEC format.
func fmtSize(kilobytes int) string {
	return utils.Kilobytes(kilobytes).ToIEC()
}
