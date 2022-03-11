package configuration

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"returntypes-langserver/common/dataformat/jsonschema"
	"returntypes-langserver/common/debug/errors"
)

const ConfigurationErrorTitle = "Configuration Error"

type configFile struct {
	// Configurations for cloning
	Cloner ClonerConfiguration `json:"cloner"`
	// Defines different project configurations. The value might be:
	// - an array containing each project configurations.
	// - a string pointing to a different file which contains the project configurations.
	// The array may define a project by a detailed object or a simple string. If it is a string,
	// then it is considered as the project's git uri.
	Projects ProjectConfiguration `json:"projects"`
	// The main output dir containing some processing results and the final dataset
	MainOutputDir string `json:"mainOutputDir"`
	// A list of mapping files in csv format for default java libraries (and other libraries to use)
	DefaultLibraries []string `json:"defaultLibraries"`
	// A file containg type mappings and the final labels
	DefaultTypeClasses string `json:"defaultTypeClasses"`
	// Configuration for the crawler
	Crawler CrawlerConfiguration `json:"crawler"`
	// If true, will always recollect the data from the crawled xml files
	ForceExtraction bool `json:"forceExtraction"`
	// Defines for which model type the dataset should be generated / which model type should be trained
	ModelType ModelType `json:"modelType"`
	// Configurations for the predictor
	Predictor PredictorConfiguration `json:"predictor"`
	// Activates strict mode.
	// If activated, the program will stop running when an error occurs which may influence the dataset output
	StrictMode bool `json:"strictMode"`
	// Configurations for logging
	Logger LoggerConfiguration `json:"logger"`
	// Configurations for connections
	Connection ConnectionConfiguration `json:"connection"`
	// Configurations for statistics
	Statistics StatisticsConfiguration `json:"statistics"`
	// The excel set configurations.
	ExcelSets ExcelSetConfiguration `json:"excelSets"`
	// Evaluation configurations
	Evaluation EvaluationConfiguration `json:"evaluation"`
	// Dataset configurations
	Datasets DatasetConfiguration `json:"datasets"`
	// If true, then the program is in language server mode (command line only)
	IsLangServMode bool
}

// The proportions for the size of datasets
type DatasetProportion struct {
	Training   float64 `json:"training"`
	Evaluation float64 `json:"evaluation"`
}

type ClonerConfiguration struct {
	// If true, uses the command line git tool to clone projects. This allows the use of some options (like "--filter") which speed up the clone process.
	// "git" must be available on PATH to use this.
	UseCommandLineTool bool `json:"useCommandLineTool"`
	// Maximum size in kilobytes of the (github) repositories to clone. If exceeded, clone process will be skipped
	MaximumCloneSize int `json:"maximumCloneSize"`
	// If true, skip clone process
	Skip bool `json:"skip"`
	// The directory where projects will be cloned into
	OutputDir string `json:"outputDir"`
}

type CrawlerConfiguration struct {
	// The path to the .jar file of the crawler version which should be used
	ExecutablePath string `json:"executablePath"`
	// The default java version the crawler should use to parse Java files (if not overwritten by project settings)
	// If set to zero, then it is left to the parser library to decide which version should be used.
	DefaultJavaVersion int `json:"defaultJavaVersion"`
}

type PredictorConfiguration struct {
	// The path to the script starting the python scripts
	ScriptPath string `json:"scriptPath"`
	// The host of the predictor
	Host string `json:"host"`
	// The port the predictor listens to for predictions
	Port int `json:"port"`
	// If true, the training process will be skipped
	SkipTraining bool `json:"skipTraining"`
	// If true, uses the mocked predictor implementation
	UseMock bool `json:"useMock"`
}

type LoggerConfiguration struct {
	// The port the remote debug logger listens to
	RemotePort int `json:"port"`
	// If true, activates logging using the remote debugger in language server mode.
	ActivateRemoteLogging bool `json:"activateRemoteLogging"`
	// The layers which may be logged (See logger.go for a list of possible values)
	// If nil, only the layers "critical" and "information" will be logged.
	Layers []string `json:"layers"`
	// If true, errors are also fully visible in console output. (If the critical logging layer is not active,
	// errors won't be logged anywhere)
	// Otherwise, only a brief explanation of the error is shown. (The full error will be still written to the logfile if configured)
	ErrorsInConsoleOutput bool `json:"errorsInConsoleOutput"`
}

type ConnectionConfiguration struct {
	// The time to wait in ms after a failing connection attempt
	Timeout int `json:"timeout"`
	// The amount of attempts for trying to reconnect
	ReconnectionAttempts int `json:"reconnectionAttempts"`
}

type StatisticsConfiguration struct {
	// If true, the statistics creation will be skipped
	SkipCreation bool `json:"skip"`
	// The minimum amount of occurences for methods to be visible in the Methods before summarization table
	// If its one, all methods are shown (not recommended)
	MinOccurencesForMethodsBeforeSummarizationTable int `json:"minOccurencesForMethodsBeforeSumarizationTable"`
	// All projects which's value is below this value will be grouped as a "other projects" value.
	// This does only affect the "Origins of methods used in the final dataset" pie chart.
	ProjectGroupingThreshold float64 `json:"projectGroupingThreshold"`
}

var loadedConfig *configFile

// Needs to be called in order to read the configuration's values
func Load() errors.Error {
	createDefaultConfig()
	initCommandLineArguments()
	err := loadConfigFromFile()
	loadCommandLineArguments()

	return err
}

// Loads configurations from a json string. Only the setted values will overwrite the current configuration.
func LoadConfigFromJsonString(jsonStr string) errors.Error {
	if loadedConfig == nil {
		createDefaultConfig()
	}
	return loadJsonConfig([]byte(jsonStr))
}

func createDefaultConfig() {
	loadedConfig = &configFile{
		Cloner: ClonerConfiguration{
			UseCommandLineTool: false,
			MaximumCloneSize:   1024 * 512,
			Skip:               false,
			OutputDir:          "",
		},
		MainOutputDir:    filepath.Join(GoProjectDir(), "results"),
		DefaultLibraries: []string{filepath.Join(GoProjectDir(), "resources", "data", "javalang.csv")},
		Crawler: CrawlerConfiguration{
			ExecutablePath:     filepath.Join(GoProjectDir(), "resources", "crawler", "returntypes-crawler.jar"),
			DefaultJavaVersion: 0,
		},
		ForceExtraction: false,
		ModelType:       MethodGenerator,
		Predictor: PredictorConfiguration{
			Port:         10000,
			Host:         "localhost",
			ScriptPath:   "",
			SkipTraining: false,
			UseMock:      false,
		},
		StrictMode: false,
		Logger: LoggerConfiguration{
			RemotePort:            9000,
			ActivateRemoteLogging: false,
			Layers:                nil,
			ErrorsInConsoleOutput: false,
		},
		Connection: ConnectionConfiguration{
			Timeout:              10000,
			ReconnectionAttempts: 5,
		},
		Statistics: StatisticsConfiguration{
			SkipCreation: false,
			MinOccurencesForMethodsBeforeSummarizationTable: 50,
			ProjectGroupingThreshold:                        0.01,
		},
		IsLangServMode: false,
	}
}

func loadConfigFromFile() errors.Error {
	file, err := os.Open(filepath.Join(GoProjectDir(), "config.json"))
	if err != nil {
		// No configuration specified.
		return nil
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return errors.Wrap(err, ConfigurationErrorTitle, "Could not load the configuration file")
	}
	return loadJsonConfig(content)
}

func loadJsonConfig(content []byte) errors.Error {
	initializeSchemas()
	if err := jsonschema.UnmarshalJSONStrict(content, &loadedConfig, ProjectConfigurationFileSchema); err != nil {
		return errors.Wrap(err, ConfigurationErrorTitle, "Could not load json configuration")
	}
	connectDatasetPaths(loadedConfig.Datasets, "")
	return nil
}

// Sets the program in language server mode
func SetLangServMode(state bool) {
	if loadedConfig != nil {
		loadedConfig.IsLangServMode = state
	}
}
