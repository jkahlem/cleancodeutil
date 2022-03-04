package configuration

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"returntypes-langserver/common/utils"
	"strings"
)

type ProjectConfiguration []Project

type projectConfigurationFileStructure struct {
	Projects ProjectConfiguration `json:"projects"`
}

func (c *ProjectConfiguration) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	if filePath, ok := v.(string); ok {
		// Load configuration from different JSON file
		return c.fromFilePath(filePath)
	} else if slice, ok := v.([]interface{}); ok {
		return c.fromSlice(slice)
	} else {
		return fmt.Errorf("Unsupported project configuration value: %v", v)
	}
}

func (c *ProjectConfiguration) fromFilePath(filePath string) error {
	contents, err := ioutil.ReadFile(AbsolutePathFromGoProjectDir(filePath))
	if err != nil {
		return err
	}
	var fileConfig projectConfigurationFileStructure
	if err := json.Unmarshal(contents, &fileConfig); err != nil {
		return err
	}
	*c = fileConfig.Projects
	return nil
}

func (c *ProjectConfiguration) fromSlice(slice []interface{}) error {
	*c = make(ProjectConfiguration, len(slice))
	for i, element := range slice {
		if err := (*c)[i].fromInterface(element); err != nil {
			return err
		}
	}
	return nil
}

type Project struct {
	// If set and the project is currently not cloned then the repository will be cloned from this URI.
	GitUri string `json:"gitUri"`
	// Sets the directory where the project should be loaded from. If a git uri is set and the directory does not exist
	// on the file system, the project will be cloned here.
	// If no directory is set, the project's directory will be {projectInputDir}/{repositoryName}
	Directory string `json:"directory"`
	// Sets an alternative name for the repository which is usefull if two different repositories have the same name.
	// If set and the directory attribute is empty, the project's directory will be {projectInputDir}/{alternativeName}.
	AlternativeName string `json:"alternativeName"`
	// Sets the java version to be used for parsing the project's source code.
	JavaVersion int `json:"javaVersion"`
}

func (c *Project) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	return c.fromInterface(v)
}

func (c *Project) fromInterface(itf interface{}) error {
	if uri, ok := itf.(string); ok {
		c.GitUri = uri
	} else if jsonObj, ok := itf.(map[string]interface{}); ok {
		if err := utils.DecodeMapToStructStrict(jsonObj, c); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Unsupported project configuration value: %v", itf)
	}

	c.GitUri = strings.TrimRight(c.GitUri, "/")
	return nil
}
