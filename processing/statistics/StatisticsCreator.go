package statistics

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
	"returntypes-langserver/common/utils"
	"returntypes-langserver/processing/dataset/base"
	"returntypes-langserver/processing/git"
	"returntypes-langserver/processing/projects"
	"returntypes-langserver/services/predictor"
)

const StatisticsErrorTitle = "Statistics Error"
const UnknownType string = "unknown"

// Creates statistics
type StatisticsCreator struct {
	typeLabelMapper *base.TypeLabelMapper
	builder         StatisticsBuilder
}

// Contains needed repository information
type RepositoryInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Creates statistics and writes them to the output files
func CreateStatistics(projects []projects.Project) errors.Error {
	log.Info("Create statistics for the dataset...\n")
	creator := StatisticsCreator{}
	return creator.Create(projects)
}

// Creates statistics and writes them to the output files
func (c *StatisticsCreator) Create(projects []projects.Project) errors.Error {
	// prepare creator
	if err := c.prepare(); err != nil {
		return err
	}

	// create statistics
	stats, err := c.createStatistics(projects)
	if err != nil {
		return err
	}

	// create charts
	CreateCharts(stats)

	// write data as json
	if err := c.writeStatisticsAsJson(stats); err != nil {
		return err
	}
	return nil
}

func (c *StatisticsCreator) prepare() errors.Error {
	c.typeLabelMapper = &base.TypeLabelMapper{}
	if err := c.typeLabelMapper.LoadFromFile(configuration.DatasetLabelsOutputPath()); err != nil {
		return err
	}
	return nil
}

// Creates the statistics
func (c *StatisticsCreator) createStatistics(projects []projects.Project) (Statistics, errors.Error) {
	if err := c.addProjectInfos(projects); err != nil {
		return Statistics{}, err
	} else if err = c.addFileCounts(projects); err != nil {
		return Statistics{}, err
	} else if err = c.addMethods(); err != nil {
		return Statistics{}, err
	}
	// it is allowed that there is no evaluation result/summarized methods data
	c.addEvaluationResult()
	c.addSummarizedMethodsData()
	return c.builder.Build(), nil
}

// Adds project info to the statistics
func (c *StatisticsCreator) addProjectInfos(projects []projects.Project) errors.Error {
	for _, project := range projects {
		c.addProjectInfo(project.Name())
	}
	return nil
}

// Adds project info (name and description) of the given repository/project
func (c *StatisticsCreator) addProjectInfo(name string) {
	if info, successful := c.loadRepositoryInfo(name); successful {
		c.builder.AddProjectInfo(name, info.Name, info.Description)
	}
}

// Loads the repository info of the repository/project with the given name
func (c *StatisticsCreator) loadRepositoryInfo(name string) (RepositoryInfo, bool) {
	info := RepositoryInfo{}
	if content, err := git.LoadRepositoryInfo(name); err != nil {
		return info, false
	} else if err := json.Unmarshal(content, &info); err != nil {
		return info, false
	}
	return info, true
}

// Adds the file count statistics
func (c *StatisticsCreator) addFileCounts(projects []projects.Project) errors.Error {
	for _, project := range projects {
		path := filepath.Join(configuration.CrawlerOutputDir(), project.Name())
		if !utils.FileExists(path) {
			continue
		} else if nodeCount, err := c.getFileNodesCountOfXmlFile(path); err != nil {
			return err
		} else {
			c.builder.AddFileCount(project.Name(), nodeCount)
		}
	}
	return nil
}

// Returns the amount of java code file nodes in the crawler output xml file
func (c *StatisticsCreator) getFileNodesCountOfXmlFile(path string) (int, errors.Error) {
	if xmlobj, err := loadOnlyFileNodesFromXML(path); err != nil {
		return 0, err
	} else {
		return len(xmlobj.Files), nil
	}
}

// Adds statistics for methods
func (c *StatisticsCreator) addMethods() errors.Error {
	if err := c.addDatasetMethods(); err != nil {
		return err
	} else if err := c.addAllExtractedMethods(); err != nil {
		return err
	}
	return nil
}

// Adds statistics using the methods which finally got into the dataset.
func (c *StatisticsCreator) addDatasetMethods() errors.Error {
	if err := c.addDatasetMethodsFromDataset(configuration.TrainingSetOutputPath()); err != nil {
		return err
	} else if err = c.addDatasetMethodsFromDataset(configuration.EvaluationSetOutputPath()); err != nil {
		return err
	}
	return nil
}

// Adds method statistics for dataset methods of the given dataset
func (c *StatisticsCreator) addDatasetMethodsFromDataset(path string) errors.Error {
	rows, err := c.loadDatasetRows(path)
	if err != nil {
		return err
	}
	for _, row := range rows {
		typeName := c.convertLabelToTypeName(row.TypeLabel)
		c.builder.AddDatasetMethod(predictor.PredictableMethodName(row.MethodName), typeName)
	}
	return nil
}

// Converts a dataset type label to the corresponding type name (using the created type labels map)
func (c *StatisticsCreator) convertLabelToTypeName(typeLabel int) string {
	if c.typeLabelMapper != nil {
		if typeName, ok := c.typeLabelMapper.GetTypeName(typeLabel); ok {
			return typeName
		}
	}
	return UnknownType
}

// Loads and unmarshals rows of the given dataset
func (c *StatisticsCreator) loadDatasetRows(path string) ([]csv.ReturnTypesDatasetRow, errors.Error) {
	records, err := csv.NewFileReader(path).ReadReturnTypesDatasetRowRecords()
	if err != nil {
		return nil, err
	}
	return records, nil
}

// Adds statistics using the methods which were extracted from the project before filtering them for the dataset.
func (c *StatisticsCreator) addAllExtractedMethods() errors.Error {
	if methods, err := c.loadAllExtractedMethods(); err != nil {
		return err
	} else {
		for _, method := range methods {
			c.addExtractedMethod(method)
		}
		return nil
	}
}

// Adds statistics for one extracted method
func (c *StatisticsCreator) addExtractedMethod(method csv.Method) {
	projectId := c.parseProjectIdFromFilepath(method.FilePath)
	c.builder.AddMethod(projectId, method.MethodName, method.ReturnType, method.Labels)
}

// Returns the project id by reading it from the path of the code file
func (c *StatisticsCreator) parseProjectIdFromFilepath(path string) string {
	return strings.Split(path, string(filepath.Separator))[0]
}

// Loads all extracted methods
func (c *StatisticsCreator) loadAllExtractedMethods() ([]csv.Method, errors.Error) {
	records, err := csv.NewFileReader(configuration.MethodsWithTypeClassesOutputPath()).ReadMethodRecords()
	if err != nil {
		return nil, err
	}
	return records, nil
}

// Adds the evaluation result to the statistics
func (c *StatisticsCreator) addEvaluationResult() {
	if evaluationResult, err := c.loadEvaluationResult(); err != nil {
		return
	} else {
		c.builder.AddEvaluationResult(evaluationResult)
	}
}

// Loads the evaluation result
func (c *StatisticsCreator) loadEvaluationResult() (predictor.Evaluation, errors.Error) {
	file, err := os.Open(configuration.EvaluationResultOutputPath())
	if err != nil {
		return predictor.Evaluation{}, errors.Wrap(err, "Error", "Could not load evaluation result")
	}
	defer file.Close()
	evaluationResult := predictor.Evaluation{}
	if err := json.NewDecoder(file).Decode(&evaluationResult); err != nil {
		return predictor.Evaluation{}, errors.Wrap(err, "Error", "Could not load evaluation result")
	}
	return evaluationResult, nil
}

// Adds the summarized methods data
func (c *StatisticsCreator) addSummarizedMethodsData() {
	if records, err := csv.NewFileReader(configuration.MethodSummarizationDataOutputPath()).ReadMethodSummarizationDataRecords(); err != nil {
		log.Error(err)
		return
	} else {
		c.builder.AddSummarizedMethodsData(records)
	}
}

// Writes the statistics to a json file
func (c *StatisticsCreator) writeStatisticsAsJson(stats Statistics) errors.Error {
	jsonFile, err := c.createJsonOutputFile()
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	if err := json.NewEncoder(jsonFile).Encode(stats); err != nil {
		return errors.Wrap(err, StatisticsErrorTitle, "Could not create statistics output")
	}
	return nil
}

// Creates a json file for writing statistics
func (c *StatisticsCreator) createJsonOutputFile() (*os.File, errors.Error) {
	err := os.MkdirAll(configuration.StatisticsOutputDir(), os.ModePerm)
	if err != nil {
		return nil, errors.Wrap(err, StatisticsErrorTitle, "Could not create statistics output")
	}
	jsonFile, err := os.Create(configuration.RawStatisticsOutputPath())
	if err != nil {
		return nil, errors.Wrap(err, StatisticsErrorTitle, "Could not create statistics output")
	}
	return jsonFile, nil
}
