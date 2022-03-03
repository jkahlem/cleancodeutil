// The configuration package allows loading the programs configuration from different
// sources (e.g. command line, config.json file etc.) and exports functions for reading
// values of each of the current configurations.
//
// If no configuration is loaded, the functions will return an undefined value but will
// never panic.
package configuration

import (
	"path/filepath"
	"time"
)

func Projects() []Project {
	if loadedConfig == nil {
		return nil
	}
	return loadedConfig.Projects
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

func CrawlerPath() string {
	if loadedConfig == nil {
		return ""
	}
	return AbsolutePathFromGoProjectDir(loadedConfig.CrawlerPath)
}

func ForceExtraction() bool {
	if loadedConfig == nil {
		return false
	}
	return loadedConfig.ForceExtraction
}

func DatasetSize() DatasetProportion {
	if loadedConfig == nil {
		return DatasetProportion{
			Training:   0,
			Evaluation: 0,
		}
	}
	return loadedConfig.DatasetSize
}

func TrainingSetSize() float64 {
	if loadedConfig == nil {
		return 0
	}
	return loadedConfig.DatasetSize.Training
}

func EvaluationSetSize() float64 {
	if loadedConfig == nil {
		return 0
	}
	return loadedConfig.DatasetSize.Evaluation
}

func MethodFilter() DatasetMethodFilter {
	if loadedConfig == nil {
		return DatasetMethodFilter{}
	}
	return loadedConfig.MethodFilter
}

func MethodFilterGetter() bool {
	if loadedConfig == nil {
		return false
	}
	return loadedConfig.MethodFilter.Getter
}

func MethodFilterSetter() bool {
	if loadedConfig == nil {
		return false
	}
	return loadedConfig.MethodFilter.Setter
}

func MethodFilterOverride() bool {
	if loadedConfig == nil {
		return false
	}
	return loadedConfig.MethodFilter.Override
}

func MethodFilterTestCode() bool {
	if loadedConfig == nil {
		return false
	}
	return loadedConfig.MethodFilter.TestCode
}

func PredictorScriptPath() string {
	if loadedConfig == nil {
		return ""
	}
	return AbsolutePathFromGoProjectDir(loadedConfig.Predictor.ScriptPath)
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

func StatisticsSkipCreation() bool {
	if loadedConfig == nil {
		return false
	}
	return loadedConfig.Statistics.SkipCreation
}

func StatisticsMinOccurencesForMethodsBeforeSummarizationTable() int {
	if loadedConfig == nil {
		return 0
	}
	return loadedConfig.Statistics.MinOccurencesForMethodsBeforeSummarizationTable
}

func StatisticsProjectGroupingThreshold() float64 {
	if loadedConfig == nil {
		return 0
	}
	return loadedConfig.Statistics.ProjectGroupingThreshold
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

// The path where files for statistics will be saved
func StatisticsOutputDir() string {
	return filepath.Join(MainOutputDir(), "statistics")
}

// The path to the file containing raw statistics output in JSON-format
func RawStatisticsOutputPath() string {
	return filepath.Join(StatisticsOutputDir(), "statistics.json")
}

// The path to the file containing charts for the statistics
func ChartsOutputPath() string {
	return filepath.Join(StatisticsOutputDir(), "charts.html")
}

// The path to the methods file containing all methods with the return types as type classes
func MethodsWithTypeClassesOutputPath() string {
	return filepath.Join(StatisticsOutputDir(), "methodsWithTypeClasses.csv")
}

// The path where the evaluation result is saved for the statistics
func EvaluationResultOutputPath() string {
	return filepath.Join(StatisticsOutputDir(), "evaluationResult.json")
}

// The path where the method summarization data (a filtered method list with return type counts) are saved.
// This data contains the methods which are filtered using labels etc. before methods with the same name(/sentence) are summarized to one dataset row.
// This data is the basis for the summarization process and will be used for the statistics.
func MethodSummarizationDataOutputPath() string {
	return filepath.Join(StatisticsOutputDir(), "methodSummarizationData.csv")
}

// The directory where the excel output is saved
func MethodsWithReturnTypesExcelOutputDir() string {
	return filepath.Join(StatisticsOutputDir(), "excel")
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
