package statistics

import (
	"returntypes-langserver/common/code/java"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/services/predictor"
)

// Builds statistics.
type StatisticsBuilder struct {
	datasetMethodMap DatasetMethodMap
	projects         map[string]*ProjectStatistics
	general          GeneralStatistics
	dataset          SpecificStatistics
	testcode         SpecificStatistics
	labels           LabelStatistics
	evaluation       *predictor.Evaluation
}

type MethodEntry struct {
	returnType string
	projects   []string
}

type DatasetMethodMap map[predictor.PredictableMethodName]*MethodEntry

// Adds project infos to statistics
func (builder *StatisticsBuilder) AddProjectInfo(projectId, name, description string) {
	builder.getProject(projectId).Name = name
	builder.getProject(projectId).Description = description
}

// Adds file count to statistics
func (builder *StatisticsBuilder) AddFileCount(projectId string, count int) {
	builder.getProject(projectId).FileCount += count
	builder.general.FileCount += count
}

// Adds a method belonging to the final dataset to the statistics.
func (builder *StatisticsBuilder) AddDatasetMethod(methodName predictor.PredictableMethodName, returnType string) {
	builder.dataset.MethodsCount++
	builder.dataset.ReturnTypes.AddUsage(returnType)
	builder.addDatasetMethodToMap(methodName, returnType)
}

func (builder *StatisticsBuilder) addDatasetMethodToMap(methodName predictor.PredictableMethodName, returnType string) {
	if builder.datasetMethodMap == nil {
		builder.datasetMethodMap = make(DatasetMethodMap)
	}
	builder.datasetMethodMap[methodName] = &MethodEntry{
		returnType: returnType,
		projects:   make([]string, 2),
	}
}

// Adds a method which belongs to the code of the repositories to the statistics.
func (builder *StatisticsBuilder) AddMethod(projectId, methodName, returnType string, labels []string) {
	builder.general.MethodsCount++
	builder.general.ReturnTypes.AddUsage(returnType)
	builder.getProject(projectId).MethodsCount++
	builder.getProject(projectId).ReturnTypes.AddUsage(returnType)
	if builder.isMethodInDataset(projectId, methodName) {
		builder.getProject(projectId).MethodsInDatasetCount++
	}
	builder.AddLabels(labels, returnType)
}

// Returns true if the method is inside the output dataset for only the first check for the given project
func (builder *StatisticsBuilder) isMethodInDataset(projectId, methodName string) bool {
	if builder.datasetMethodMap != nil {
		predictableName := predictor.GetPredictableMethodName(methodName)
		if method, ok := builder.datasetMethodMap[predictableName]; ok {
			// check if this method was already counted for this project
			for _, project := range method.projects {
				if project == projectId {
					return false
				}
			}
			method.projects = append(method.projects, projectId)
			return true
		}
	}
	return false
}

// Adds label usages for a specific type to the statistics
func (builder *StatisticsBuilder) AddLabels(labels []string, returnType string) {
	for _, label := range labels {
		switch java.MethodLabel(label) {
		case java.Getter:
			builder.labels.Getter++
		case java.Setter:
			builder.labels.Setter++
		case java.ArrayType:
			builder.labels.ArrayType++
		case java.Override:
			builder.labels.Override++
		case java.ChainMethod:
			builder.labels.ChainMethod++
		case java.TestCode:
			builder.labels.TestCode++
			builder.testcode.MethodsCount++
			builder.testcode.ReturnTypes.AddUsage(returnType)
		}
	}
}

// Adds the training evaluation result to the statistics
func (builder *StatisticsBuilder) AddEvaluationResult(evaluationResult predictor.Evaluation) {
	builder.evaluation = &evaluationResult
}

// Adds sumarized methods data to the statistics
func (builder *StatisticsBuilder) AddSummarizedMethodsData(data []csv.MethodSummarizationData) {
	builder.general.MethodListBeforeSummarization = data
}

// Returns the statistics for a given project
func (builder *StatisticsBuilder) getProject(id string) *ProjectStatistics {
	if builder.projects == nil {
		builder.projects = make(map[string]*ProjectStatistics)
	}
	if p, ok := builder.projects[id]; ok {
		return p
	} else {
		p := &ProjectStatistics{DirName: id, Name: id}
		builder.projects[id] = p
		return p
	}
}

// Builds the statistics.
func (builder *StatisticsBuilder) Build() Statistics {
	stats := Statistics{
		General:    builder.general,
		Dataset:    builder.dataset,
		TestCode:   builder.testcode,
		Labels:     builder.labels,
		Evaluation: builder.evaluation,
	}
	stats.Projects = make([]ProjectStatistics, 0, len(builder.projects))
	for _, project := range builder.projects {
		stats.Projects = append(stats.Projects, *project)
	}
	return stats
}
