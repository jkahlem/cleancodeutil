package configuration

import "returntypes-langserver/common/dataformat/jsonschema"

//go:generate go run ../dataformat/jsonschema/schemaToCode ../../schemas

const (
	// Global configuration schemas
	ConfigurationSchemaPath               = "configuration/configuration.schema.json"
	ClonerConfigurationSchemaPath         = "configuration/cloner-configuration.schema.json"
	CrawlerConfigurationSchemaPath        = "configuration/crawler-configuration.schema.json"
	ConnectionsConfigurationSchemaPath    = "configuration/connections-configuration.schema.json"
	LoggerConfigurationSchemaPath         = "configuration/logger-configuration.schema.json"
	PredictorConfigurationSchemaPath      = "configuration/predictor-configuration.schema.json"
	StatisticsConfigurationSchemaPath     = "configuration/statistics-configuration.schema.json"
	LanguageServerConfigurationSchemaPath = "configuration/language-server-configuration.schema.json"

	// Excel set schemas
	FilterSchemaPath                    = "datasets/filter.schema.json"
	FilterConfigurationSchemaPath       = "datasets/filter-configuration.schema.json"
	PatternSchemaPath                   = "datasets/pattern.schema.json"
	ExcelSetConfigurationFileSchemaPath = "datasets/excel-set-config-file.schema.json"
	ExcelSetConfigurationSchemaPath     = "datasets/excel-set-configuration.schema.json"

	// Project configuration schemas
	ProjectConfigurationFileSchemaPath = "projects/project-config-file.schema.json"
	ProjectConfigurationSchemaPath     = "projects/project-definition.schema.json"

	// Evaluation schemas
	EvaluationConfigurationSchemaPath = "configuration/evaluation-configuration.schema.json"
	EvaluationSetSchemaPath           = "datasets/evaluation-set.schema.json"
	MethodContextSchemaPath           = "datasets/method-context.schema.json"

	// Dataset schemas
	DatasetConfigurationFileSchemaPath    = "datasets/dataset/config-file.schema.json"
	DatasetConfigurationBaseSchemaPath    = "datasets/dataset/base.schema.json"
	DatasetConfigurationSchemaPath        = "datasets/dataset/configuration.schema.json"
	DatasetModelOptionsSchemaPath         = "datasets/dataset/model-options.schema.json"
	DatasetCreationOptionsSchemaPath      = "datasets/dataset/creation-options.schema.json"
	DatasetPreprocessingOptionsSchemaPath = "datasets/dataset/preprocessing-options.schema.json"
	DatasetSizeSchemaPath                 = "datasets/dataset/dataset-size.schema.json"
	AdamConfigurationSchemaPath           = "datasets/dataset/adam.schema.json"
	AdafactorConfigurationSchemaPath      = "datasets/dataset/adafactor.schema.json"

	// Type class schemas
	TypeClassConfigurationFileSchemaPath = "typeclasses/typeclass-config-file.schema.json"
	TypeClassSchemaPath                  = "typeclasses/typeclass.schema.json"

	// Metrics
	MetricsSchemaPath  = "metrics/metrics.schema.json"
	BleuSchemaPath     = "metrics/bleu.schema.json"
	RougeLSchemaPath   = "metrics/rouge-l.schema.json"
	RougeNSchemaPath   = "metrics/rouge-n.schema.json"
	RougeSSchemaPath   = "metrics/rouge-s.schema.json"
	MeasuresSchemaPath = "metrics/measures.schema.json"
	FscoreSchemaPath   = "metrics/fscore.schema.json"

	// Model list schema
	ModelListSchemaPath = "datasets/model-list.schema.json"
)

var ExcelSetConfigurationFileSchema,
	ProjectConfigurationFileSchema,
	EvaluationConfigurationFileSchema,
	DatasetConfigurationFileSchema,
	TypeClassConfigurationFileSchema,
	ConfigurationFileSchema jsonschema.Schema

func initializeSchemas() {
	ExcelSetConfigurationFileSchema = jsonschema.FromMap(getSchemaMap()).
		WithTopLevel(ExcelSetConfigurationFileSchemaPath).
		WithResources(ExcelSetConfigurationSchemaPath,
			FilterSchemaPath,
			FilterConfigurationSchemaPath,
			PatternSchemaPath).
		MustCompile()

	ProjectConfigurationFileSchema = jsonschema.FromMap(getSchemaMap()).
		WithTopLevel(ProjectConfigurationFileSchemaPath).
		WithResources(ProjectConfigurationSchemaPath).
		MustCompile()

	EvaluationConfigurationFileSchema = jsonschema.FromMap(getSchemaMap()).
		WithTopLevel(EvaluationConfigurationSchemaPath).
		WithResources(EvaluationSetSchemaPath,
			MethodContextSchemaPath,
			ModelListSchemaPath,
			FilterSchemaPath,
			FilterConfigurationSchemaPath,
			PatternSchemaPath,
			MetricsSchemaPath,
			BleuSchemaPath,
			RougeLSchemaPath,
			RougeNSchemaPath,
			RougeSSchemaPath,
			MeasuresSchemaPath,
			FscoreSchemaPath).
		MustCompile()

	DatasetConfigurationFileSchema = jsonschema.FromMap(getSchemaMap()).
		WithTopLevel(DatasetConfigurationFileSchemaPath).
		WithResources(DatasetConfigurationBaseSchemaPath,
			DatasetConfigurationSchemaPath,
			DatasetModelOptionsSchemaPath,
			DatasetPreprocessingOptionsSchemaPath,
			DatasetCreationOptionsSchemaPath,
			DatasetSizeSchemaPath,
			AdafactorConfigurationSchemaPath,
			AdamConfigurationSchemaPath,
			TypeClassSchemaPath,
			ModelListSchemaPath,
			FilterSchemaPath,
			FilterConfigurationSchemaPath,
			PatternSchemaPath).
		MustCompile()

	TypeClassConfigurationFileSchema = jsonschema.FromMap(getSchemaMap()).
		WithTopLevel(TypeClassConfigurationFileSchemaPath).
		WithResources(TypeClassSchemaPath).
		MustCompile()

	ConfigurationFileSchema = jsonschema.FromMap(getSchemaMap()).
		WithTopLevel(ConfigurationSchemaPath).
		WithResources(ClonerConfigurationSchemaPath,
			CrawlerConfigurationSchemaPath,
			ConnectionsConfigurationSchemaPath,
			LoggerConfigurationSchemaPath,
			PredictorConfigurationSchemaPath,
			StatisticsConfigurationSchemaPath,
			LanguageServerConfigurationSchemaPath,
			ExcelSetConfigurationSchemaPath,
			FilterSchemaPath,
			FilterConfigurationSchemaPath,
			PatternSchemaPath,
			ProjectConfigurationSchemaPath,
			EvaluationConfigurationSchemaPath,
			EvaluationSetSchemaPath,
			MetricsSchemaPath,
			BleuSchemaPath,
			RougeLSchemaPath,
			RougeNSchemaPath,
			RougeSSchemaPath,
			MeasuresSchemaPath,
			FscoreSchemaPath,
			MethodContextSchemaPath,
			DatasetConfigurationBaseSchemaPath,
			DatasetConfigurationSchemaPath,
			DatasetModelOptionsSchemaPath,
			DatasetCreationOptionsSchemaPath,
			DatasetPreprocessingOptionsSchemaPath,
			DatasetSizeSchemaPath,
			AdafactorConfigurationSchemaPath,
			AdamConfigurationSchemaPath,
			TypeClassSchemaPath,
			ModelListSchemaPath).
		MustCompile()
}
