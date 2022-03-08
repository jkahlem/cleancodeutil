package configuration

import "returntypes-langserver/common/dataformat/jsonschema"

var SchemaRoot = GoProjectDir() + "/schemas"

const (
	// Global configuration schemas
	ConfigurationSchemaPath            = "configuration/configuration.schema.json"
	ClonerConfigurationSchemaPath      = "configuration/cloner-configuration.schema.json"
	CrawlerConfigurationSchemaPath     = "configuration/crawler-configuration.schema.json"
	ConnectionsConfigurationSchemaPath = "configuration/connections-configuration.schema.json"
	LoggerConfigurationSchemaPath      = "configuration/logger-configuration.schema.json"
	PredictorConfigurationSchemaPath   = "configuration/predictor-configuration.schema.json"
	StatisticsConfigurationSchemaPath  = "configuration/statistics-configuration.schema.json"

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
)

// TODO: Find a way to not use must compile?
var ExcelSetConfigurationFileSchema = jsonschema.AtRoot(SchemaRoot).
	WithTopLevel(ExcelSetConfigurationFileSchemaPath).
	WithResources(FilterSchemaPath, FilterConfigurationSchemaPath, PatternSchemaPath).
	MustCompile()

var ProjectConfigurationFileSchema = jsonschema.AtRoot(SchemaRoot).
	WithTopLevel(ProjectConfigurationFileSchemaPath).
	WithResources(ProjectConfigurationSchemaPath).
	MustCompile()

var EvaluationConfigurationFileSchema = jsonschema.AtRoot(SchemaRoot).
	WithTopLevel(EvaluationConfigurationSchemaPath).
	WithResources(EvaluationSetSchemaPath).
	MustCompile()

var ConfigurationFileSchema = jsonschema.AtRoot(SchemaRoot).
	WithTopLevel(ConfigurationSchemaPath).
	WithResources(ClonerConfigurationSchemaPath,
		CrawlerConfigurationSchemaPath,
		ConnectionsConfigurationSchemaPath,
		LoggerConfigurationSchemaPath,
		PredictorConfigurationSchemaPath,
		StatisticsConfigurationSchemaPath,
		ExcelSetConfigurationSchemaPath,
		FilterSchemaPath,
		FilterConfigurationSchemaPath,
		PatternSchemaPath,
		ProjectConfigurationSchemaPath,
		EvaluationConfigurationSchemaPath,
		EvaluationSetSchemaPath).
	MustCompile()
