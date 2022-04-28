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
	// If true, creates method output per project
	CreateMethodOutputPerProject bool `json:"createMethodOutputPerProject"`
	// The excel set configurations.
	ExcelSets ExcelSetConfiguration `json:"excelSets"`
	// Evaluation configurations
	Evaluation EvaluationConfiguration `json:"evaluation"`
	// Dataset configurations
	Datasets DatasetConfiguration `json:"datasets"`
	// If true, skips all processes (like dataset creation, excel set creation etc.) if the output files exists
	SkipIfOutputExists bool `json:"skipIfOutputExists"`
	// Configurations which are specific for the language server
	LanguageServer LanguageServerConfiguration `json:"languageServer"`
	// Creates statistics on preprocessed data (token counts).
	CreateStatistics bool `json:"createStatistics"`
	// Additional prefix which is added to datasets for experimental uses etc.
	DatasetPrefix string
	// If not empty, names a set for which the training process should be continued if it already exists
	// Multiple datasets are separated by a semicolon ';'.
	ContinueTraining string
	// If true, then the program is in language server mode (command line only)
	IsLangServMode bool
}

// The proportions for the size of datasets
type DatasetProportion struct {
	Training   float64 `json:"training,omitempty"`
	Evaluation float64 `json:"evaluation,omitempty"`
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
	// The host of the predictor
	Host string `json:"host"`
	// The port the predictor listens to for predictions
	Port int `json:"port"`
	// If true, the training process will be skipped
	SkipTraining bool `json:"skipTraining"`
	// If true, uses the mocked predictor implementation
	UseMock bool `json:"useMock"`
	// A list of types (simple identifiers) which will be used as default context types in predictor requests.
	DefaultContextTypes []string `json:"defaultContextTypes"`
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

type LanguageServerConfiguration struct {
	// Identifier for the dataset configuration which should be used for the language server. Can be a list splitted with slashes '/' to
	// reference subsets.
	Models LanguageServerModelConfiguration `json:"models"`
}

type LanguageServerModelConfiguration struct {
	ReturnTypesValidator string `json:"returntypesValidator"`
	MethodGenerator      string `json:"methodGenerator"`
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

// Loads configurations from a json string and panics on any errors. This is especially for tests which rely on configurations and must compile.
func MustLoadConfigFromJsonString(jsonStr string) {
	if err := LoadConfigFromJsonString(jsonStr); err != nil {
		panic(err)
	}
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
		ForceExtraction:    false,
		SkipIfOutputExists: true,
		ModelType:          MethodGenerator,
		Predictor: PredictorConfiguration{
			Port:         10000,
			Host:         "localhost",
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
		IsLangServMode: false,
	}
}

func loadConfigFromFile() errors.Error {
	initializeSchemas()
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
