// The configuration package allows loading the programs configuration from different
// sources (e.g. command line, config.json file etc.) and exports functions for reading
// values of each of the current configurations.
//
// If no configuration is loaded, the functions will return an undefined value but will
// never panic.
package configuration

import (
	"path/filepath"
	"returntypes-langserver/common/debug/errors"
	"strings"
	"time"
)

func Projects() []Project {
	if loadedConfig == nil {
		return nil
	}
	return loadedConfig.Projects
}

func EvaluationSubsets() []EvaluationSet {
	if loadedConfig == nil {
		return nil
	}
	return loadedConfig.Evaluation.Subsets
}

func Datasets() []Dataset {
	if loadedConfig == nil {
		return nil
	}
	return loadedConfig.Datasets
}

func ContinueTraining() []string {
	if loadedConfig == nil {
		return nil
	}
	return strings.Split(loadedConfig.ContinueTraining, ";")
}

var ErrDatasetNotFound = errors.ErrorId("Configuration", "Dataset not found")

func FindDatasetByReference(reference string) (Dataset, errors.Error) {
	return findDatasetByReference(strings.Split(reference, "/"), Datasets())
}

func findDatasetByReference(referenceParts []string, sets []Dataset) (Dataset, errors.Error) {
	for _, set := range sets {
		splittedName := strings.Split(referenceParts[0], ":")
		if len(splittedName) == 1 {
			// No colon: continue search for dataset
			if set.NameRaw == referenceParts[0] {
				if len(referenceParts) > 1 {
					return findDatasetByReference(referenceParts[1:], set.Subsets)
				} else {
					return set, nil
				}
			}
		} else if len(splittedName) == 2 {
			// one colon was present
			setName, alternativeName := splittedName[0], splittedName[1]
			if len(referenceParts) != 1 {
				return Dataset{}, errors.New("Configuration", "Tried to access subset of alternative set, which is not supported.")
			} else if set.NameRaw == setName {
				alternative, ok := findDatasetAlternativeByName(alternativeName, set.Alternatives)
				set.DatasetBase = alternative
				if !ok {
					return set, ErrDatasetNotFound.New()
				}
				return set, nil
			} else {
				return Dataset{}, ErrDatasetNotFound.New()
			}
		} else {
			return Dataset{}, errors.New("Configuration", "Unexpected amount of parts in reference name: `%s`. Expected 1 or 2, but got %d.", referenceParts[0], len(splittedName))
		}
	}
	return Dataset{}, ErrDatasetNotFound.New()
}

func findDatasetAlternativeByName(name string, alternatives []DatasetBase) (DatasetBase, bool) {
	for _, set := range alternatives {
		if set.NameRaw == name {
			return set, true
		}
	}
	return DatasetBase{}, false
}

func UsedModelType() ModelType {
	if loadedConfig == nil {
		return ""
	}
	return loadedConfig.ModelType
}

func DatasetPrefix() string {
	if loadedConfig == nil {
		return ""
	}
	return loadedConfig.DatasetPrefix
}

func ClonerOutputDir() string {
	if loadedConfig == nil {
		return ""
	}
	return AbsolutePathFromGoProjectDir(loadedConfig.Cloner.OutputDir)
}

func ClonerUseCommandLineTool() bool {
	if loadedConfig == nil {
		return false
	}
	return loadedConfig.Cloner.UseCommandLineTool
}

func ClonerSkip() bool {
	if loadedConfig == nil {
		return false
	}
	return loadedConfig.Cloner.Skip
}

func ClonerMaximumCloneSize() int {
	if loadedConfig == nil {
		return 1024 * 128
	}
	return loadedConfig.Cloner.MaximumCloneSize
}

func MainOutputDir() string {
	if loadedConfig == nil {
		return ""
	}
	return AbsolutePathFromGoProjectDir(loadedConfig.MainOutputDir)
}

func DefaultLibraries() []string {
	if loadedConfig == nil {
		return nil
	}
	for i, lib := range loadedConfig.DefaultLibraries {
		loadedConfig.DefaultLibraries[i] = AbsolutePathFromGoProjectDir(lib)
	}
	return loadedConfig.DefaultLibraries
}

func DefaultTypeClasses() string {
	if loadedConfig == nil {
		return ""
	}
	return AbsolutePathFromGoProjectDir(loadedConfig.DefaultTypeClasses)
}

func CrawlerExecutablePath() string {
	if loadedConfig == nil {
		return ""
	}
	return AbsolutePathFromGoProjectDir(loadedConfig.Crawler.ExecutablePath)
}

func CrawlerDefaultJavaVersion() int {
	if loadedConfig == nil {
		return 0
	}
	return loadedConfig.Crawler.DefaultJavaVersion
}

func ForceExtraction() bool {
	if loadedConfig == nil {
		return false
	}
	return loadedConfig.ForceExtraction
}

func PredictorHost() string {
	if loadedConfig == nil {
		return ""
	}
	return loadedConfig.Predictor.Host
}

func PredictorPort() int {
	if loadedConfig == nil {
		return -1
	}
	return loadedConfig.Predictor.Port
}

func PredictorSkipTraining() bool {
	if loadedConfig == nil {
		return false
	}
	return loadedConfig.Predictor.SkipTraining
}

func PredictorUseMock() bool {
	if loadedConfig == nil {
		return false
	}
	return loadedConfig.Predictor.UseMock
}

func PredictorDefaultContextTypes() []string {
	if loadedConfig == nil {
		return nil
	}
	return loadedConfig.Predictor.DefaultContextTypes
}

func StrictMode() bool {
	if loadedConfig == nil {
		return false
	}
	return loadedConfig.StrictMode
}

func LoggerRemotePort() int {
	if loadedConfig == nil {
		return -1
	}
	return loadedConfig.Logger.RemotePort
}

func LoggerActivateRemoteLogging() bool {
	if loadedConfig == nil {
		return false
	}
	return loadedConfig.Logger.ActivateRemoteLogging
}

func LoggerLayers() []string {
	if loadedConfig == nil {
		return nil
	}
	return loadedConfig.Logger.Layers
}

func LoggerErrorsInConsoleOutput() bool {
	if loadedConfig == nil {
		return false
	}
	return loadedConfig.Logger.ErrorsInConsoleOutput
}

func ConnectionTimeout() time.Duration {
	if loadedConfig == nil {
		return 3000 * time.Millisecond
	}
	return time.Duration(loadedConfig.Connection.Timeout) * time.Millisecond
}

func ConnectionReconnectionAttempts() int {
	if loadedConfig == nil {
		return 5
	}
	return loadedConfig.Connection.ReconnectionAttempts
}

func CreateMethodOutputPerProject() bool {
	if loadedConfig == nil {
		return false
	}
	return loadedConfig.CreateMethodOutputPerProject
}

func ExcelSets() []ExcelSet {
	if loadedConfig == nil {
		return nil
	}
	return loadedConfig.ExcelSets
}

func CreateStatistics() bool {
	if loadedConfig == nil {
		return false
	}
	return loadedConfig.CreateStatistics
}

func SkipIfOutputExists() bool {
	if loadedConfig == nil {
		return false
	}
	return loadedConfig.SkipIfOutputExists
}

func LanguageServerReturntypesDataset() string {
	if loadedConfig == nil {
		return ""
	}
	return loadedConfig.LanguageServer.Models.ReturnTypesValidator
}

func LanguageServerMethodGenerationDataset() string {
	if loadedConfig == nil {
		return ""
	}
	return loadedConfig.LanguageServer.Models.MethodGenerator
}

func IsLangServMode() bool {
	if loadedConfig == nil {
		return false
	}
	return loadedConfig.IsLangServMode
}

// The path the crawler's output will be saved to
func CrawlerOutputDir() string {
	return filepath.Join(MainOutputDir(), "crawler")
}

// The path the extractor's output will be saved to
func ExtractorOutputDir() string {
	return filepath.Join(MainOutputDir(), "extractor")
}

// The path to the class hierarchy file containing all classes with their extended and implemented classes
// The class hierarchy file contains NOT the classes which have no class they extend/implement (except for java.lang.Object)
func ClassHierarchyOutputPath() string {
	return filepath.Join(ExtractorOutputDir(), "classHierarchy.csv")
}

// The path to the methods file containing all methods with their return types, labels etc.
func MethodsWithReturnTypesOutputPath() string {
	return filepath.Join(ExtractorOutputDir(), "methodsWithReturnTypes.csv")
}

// The path to the file containing all context types of each java file
func FileContextTypesOutputPath() string {
	return filepath.Join(ExtractorOutputDir(), "fileContextTypes.csv")
}

// The directory where the excel output is saved
func MethodsWithReturnTypesExcelOutputDir() string {
	return filepath.Join(MainOutputDir(), "excel")
}

// The path the dataset files will be saved to
func DatasetOutputDir() string {
	return filepath.Join(MainOutputDir(), "dataset")
}

// The path the training set file will be saved to
func TrainingSetOutputPath() string {
	return filepath.Join(DatasetOutputDir(), "trainingSet.csv")
}

// The path the evaluation set file will be saved to
func EvaluationSetOutputPath() string {
	return filepath.Join(DatasetOutputDir(), "evaluationSet.csv")
}

// The path the training set file will be saved to
func MethodsTrainingSetOutputPath() string {
	return filepath.Join(DatasetOutputDir(), "trainingSetMethods.csv")
}

// The path the evaluation set file will be saved to
func MethodsEvaluationSetOutputPath() string {
	return filepath.Join(DatasetOutputDir(), "evaluationSetMethods.csv")
}

// The path the dataset labels file will be saved to
func DatasetLabelsOutputPath() string {
	return filepath.Join(DatasetOutputDir(), "datasetLabels.csv")
}

// The char used as seperator in csv files
func CsvSeperator() rune {
	// Instead of a simple comma (default setting), use semi colon to be able to use commas for lists.
	return ';'
}

// The char used as seperator for lists inside one column in csv files
func CsvListSeperator() string {
	return ","
}
