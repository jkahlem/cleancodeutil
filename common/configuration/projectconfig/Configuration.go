package projectconfig

import (
	"io/ioutil"
	"path/filepath"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/utils"
)

type ProjectConfiguration struct {
	JavaVersion int `json:"javaVersion"`
}

const ProjectConfigurationFileName = ".returntypesPredictorConfig"

func GetProjectConfiguration(directory string) (ProjectConfiguration, errors.Error) {
	contents, err := ioutil.ReadFile(filepath.Join(directory, ProjectConfigurationFileName))
	if err != nil {
		// Do nothing if the file does not exist
		return ProjectConfiguration{}, nil
	}
	var config ProjectConfiguration
	if err := utils.UnmarshalJSONStrict(contents, &config); err != nil {
		return ProjectConfiguration{}, err
	}
	return config, nil
}
