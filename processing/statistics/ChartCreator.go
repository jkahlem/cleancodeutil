package statistics

import (
	"fmt"
	"os"

	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/errors"
	"returntypes-langserver/common/log"
	"returntypes-langserver/common/utils"
	"returntypes-langserver/processing/statistics/charts"
	"returntypes-langserver/processing/typeclasses"

	echarts "github.com/go-echarts/go-echarts/v2/charts"
)

// Color code for data summarized in an "other" category
const ColorCodeForOther = "#AAAAAA"

// Creates charts.
type ChartCreator struct {
	stats     Statistics
	colorizer Colorizer
}

// Create charts for the statistics
func CreateCharts(stats Statistics) {
	creator := ChartCreator{}
	creator.Create(stats)
}

// Create charts for the statistics
func (c *ChartCreator) Create(stats Statistics) {
	c.stats = stats
	page := charts.NewPage()
	page.PageTitle = "Dataset Statistics"

	page.AddCharts(c.createTestCodeChart(),
		c.createLabelOccurenceBarChart(),
		c.createReturnTypePieChart("General", c.stats.General.ReturnTypes),
		c.createReturnTypePieChart("MainCode", c.getMainCodeReturnTypes()),
		c.createReturnTypePieChart("Dataset", c.stats.Dataset.ReturnTypes),
		c.createReturnTypePieChart("TestCode", c.stats.TestCode.ReturnTypes),
		c.createReturnTypeChartForProjects(),
		c.createProjectInfluenceChart())
	if c.stats.Evaluation != nil {
		page.AddCharts(c.createEvaluationChart())
	}
	if c.stats.General.MethodListBeforeSummarization != nil {
		page.AddCharts(c.createSummarizedMethodsChart())
	}

	// Write charts to output file
	if f, err := os.Create(configuration.ChartsOutputPath()); err != nil {
		log.Error(errors.Wrap(err, StatisticsErrorTitle, "Could not create statistics"))
	} else if err := page.Render(f); err != nil {
		log.Error(errors.Wrap(err, StatisticsErrorTitle, "Could not create statistics"))
	}
}

// Returns the return type usages in the "main code" only (so the opposite of test code)
func (c *ChartCreator) getMainCodeReturnTypes() ReturnTypeUsageCounter {
	u := ReturnTypeUsageCounter{Counts: make(map[string]int)}
	for returnType, occurences := range c.stats.General.ReturnTypes.Counts {
		u.Counts[returnType] = occurences
	}
	for returnType, occurences := range c.stats.TestCode.ReturnTypes.Counts {
		u.Counts[returnType] -= occurences
	}
	return u
}

// Creates a test code chart
func (c *ChartCreator) createTestCodeChart() *echarts.Pie {
	builder := charts.NewPieChartBuilder()
	builder.WithTitle("Methods", "in test code and not").WithDataTitle("Method count")
	builder.AddData("main", c.stats.General.MethodsCount-c.stats.TestCode.MethodsCount)
	builder.AddData("test", c.stats.TestCode.MethodsCount)
	return builder.Build()
}

// Creates a pie chart for return type usages
func (c *ChartCreator) createReturnTypePieChart(title string, data ReturnTypeUsageCounter) *echarts.Pie {
	builder := charts.NewPieChartBuilder()
	builder.WithTitle(title, "").WithDataTitle("Return type occurences")
	for returnType, occurences := range data.Counts {
		builder.AddDataWithColor(returnType, occurences, c.getReturnTypeColor(returnType))
	}
	return builder.Build()
}

// Returns the color for a given returntype
func (c *ChartCreator) getReturnTypeColor(returnType string) string {
	return c.colorizer.GetColorForTypeClass(returnType)
}

// Creates a bar chart for occurences of the method labels
func (c *ChartCreator) createLabelOccurenceBarChart() *echarts.Bar {
	builder := charts.NewBarChartBuilder()
	builder.WithTitle("Labeled methods", "").WithDataTitle("Occurences")
	builder.AddData("Getter", c.stats.Labels.Getter)
	builder.AddData("Setter", c.stats.Labels.Setter)
	builder.AddData("ArrayType", c.stats.Labels.ArrayType)
	builder.AddData("ChainMethod", c.stats.Labels.ChainMethod)
	builder.AddData("Override", c.stats.Labels.Override)
	builder.AddData("TestCode", c.stats.Labels.TestCode)
	return builder.Build()
}

// Creates a table chart for how often a return type is used in a specific project
func (c *ChartCreator) createReturnTypeChartForProjects() *charts.Table {
	table := charts.NewTable()
	table.SetTitle("Return types by projects")
	positionMap := c.createTypePositionMap()
	c.fillHeadingsFromReturnTypeMappings(table, positionMap)
	for _, project := range c.stats.Projects {
		c.addReturnTypesTableRowForProject(table, project, positionMap)
	}
	return table
}

// A "position map" for types which binds a type to a specific index (so heading and body matches in tables)
// This is needed because the return type usages for each project are not ordered and may contain not all return types.
func (c *ChartCreator) createTypePositionMap() (positionMap map[string]int) {
	typeClasses := typeclasses.GetTypeClasses()
	positionMap = make(map[string]int)
	for i, class := range typeClasses.Classes {
		positionMap[class.Label] = i + 1
	}
	return
}

// Adds headings to the table using the position map
func (c *ChartCreator) fillHeadingsFromReturnTypeMappings(table *charts.Table, positionMap map[string]int) {
	headings := make([]string, len(positionMap)+1)
	headings[0] = "Project"
	for typeLabel, index := range positionMap {
		headings[index] = typeLabel
	}
	table.SetHeadings(headings...)
	return
}

// Adds data rows to the table using the position map
func (c *ChartCreator) addReturnTypesTableRowForProject(table *charts.Table, project ProjectStatistics, positionMap map[string]int) {
	row := make([]interface{}, 1, len(positionMap)+1)
	row[0] = project.Name
	for range positionMap {
		row = append(row, " - ")
	}
	for returnType, occurences := range project.ReturnTypes.Counts {
		if index, ok := positionMap[returnType]; ok {
			row[index] = fmt.Sprintf("%d", occurences)
		}
	}
	table.AddRow(row...)
}

// Creates a influence chart of the projects. "Influence" is defined here by the amount of methods in the final dataset which belong to a given project.
// (If a project has billions of methods which are all filtered out, then also a large project can have small influences on the dataset)
func (c *ChartCreator) createProjectInfluenceChart() *echarts.Pie {
	builder := charts.NewPieChartBuilder()
	builder.WithTitle("Origins of methods used in the final dataset", "").WithDataTitle("").
		WithPercentageOnly(true).WithThreshold(configuration.StatisticsProjectGroupingThreshold(), ColorCodeForOther)
	for _, project := range c.stats.Projects {
		builder.AddData(project.Name, project.MethodsInDatasetCount)
	}
	return builder.Build()
}

// Creates a table for the training evaluation result
func (c *ChartCreator) createEvaluationChart() *charts.Table {
	table := charts.NewTable()
	table.SetTitle("Dataset evaluation result")
	table.SetHeadings("Measure", "Value")
	if c.stats.Evaluation != nil {
		table.AddRow("Acc Score", c.stats.Evaluation.AccScore)
		table.AddRow("Eval loss", c.stats.Evaluation.EvalLoss)
		table.AddRow("F1 Score", c.stats.Evaluation.F1Score)
		table.AddRow("MCC", c.stats.Evaluation.MCC)
	}
	return table
}

// Creates a table with frequently used methods and their return types occurences.
// "Accuracy" is defined by the way the dataset creation determines a type for a method which is the most-used return type of the method.
// The accuracy just shows the rate how often this is true for a given method.
func (c *ChartCreator) createSummarizedMethodsChart() *charts.Table {
	table := charts.NewTable()
	table.SetTitle("Filtered methods before summarization")
	headings := []string{"Method name", "Occurences", "Accuracy"}
	positionMap := make(map[string]int)
	length := 0

	// Create headings
	for _, method := range c.stats.General.MethodListBeforeSummarization {
		if method.Occurences < configuration.StatisticsMinOccurencesForMethodsBeforeSummarizationTable() {
			continue
		}
		for _, returnType := range method.ReturnTypes {
			if _, ok := positionMap[returnType.Name]; !ok {
				positionMap[returnType.Name] = length
				headings = append(headings, returnType.Name)
				length++
			}
		}
	}
	table.SetHeadings(headings...)

	// Create body
	for _, method := range c.stats.General.MethodListBeforeSummarization {
		if method.Occurences < configuration.StatisticsMinOccurencesForMethodsBeforeSummarizationTable() {
			continue
		}
		returnTypeCounts := make([]int, length)
		maxCount, maxIndex := -1, -1
		for _, returnType := range method.ReturnTypes {
			returnTypeCounts[positionMap[returnType.Name]] = returnType.Count
			if returnType.Count > maxCount {
				maxCount = returnType.Count
				maxIndex = positionMap[returnType.Name]
			}
		}
		acc := " - "
		if maxIndex != -1 {
			acc = fmt.Sprintf("%f", float64(returnTypeCounts[maxIndex])/float64(method.Occurences))
		}
		table.AddRow(utils.ExplodeSlices(method.Name, method.Occurences, acc, returnTypeCounts)...)
	}
	return table
}
