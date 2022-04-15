package statistics

import (
	"fmt"
	"returntypes-langserver/common/code/java"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/dataformat/excel"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/metrics"
	"returntypes-langserver/common/utils"
	"returntypes-langserver/common/utils/progressbar"
	"returntypes-langserver/services/predictor"
	"strings"

	"github.com/xuri/excelize/v2"
)

func CreateStatisticsForMethods(methods []csv.Method, outputPath string) errors.Error {
	progress := progressbar.StartNew(len(methods))
	defer progress.Finish()
	progress.SetOperation("Count sequences")

	outputSequenceTokens := TokenCount{}
	inputSequenceTokens := TokenCount{}
	fullSequenceTokens := TokenCount{}
	for _, method := range methods {
		progress.Increment()

		pars, err := java.ParseParameterList(method.Parameters)
		if err != nil {
			return err
		}
		outputSequence := getOutputSequence(pars, method.ReturnType)
		outputTokens := metrics.TokenizeSentence(predictor.SplitMethodNameToSentence(outputSequence))
		outputSequenceTokens.Add(outputTokens)

		inputSequence := getInputSequence(method)
		inputTokens := strings.Split(predictor.SplitMethodNameToSentence(inputSequence), " ")
		inputSequenceTokens.Add(inputTokens)

		fullTokens := make([]string, len(inputTokens)+len(outputTokens))
		copy(fullTokens, inputTokens)
		copy(fullTokens[len(inputTokens):], outputTokens)
		fullSequenceTokens.Add(fullTokens)
	}

	progress.SetOperation("Write output")

	file := excelize.NewFile()
	defer file.Close()

	file.Path = outputPath
	cursor := excel.NewCursor(file, "Sheet1")
	values := [][]interface{}{
		{CreateTokenChart(outputSequenceTokens, "Number of output sequences per token count")},
		{CreateTokenChart(inputSequenceTokens, "Number of input sequences per token count")},
		{CreateTokenChart(fullSequenceTokens, "Number of full sequences per token count")},
		{},
		{"Average token count in output sequences:", getAverage(outputSequenceTokens)},
		{"Average token count in input sequences:", getAverage(inputSequenceTokens)},
		{"Average token count in full sequences:", getAverage(fullSequenceTokens)},
		{},
		{"Longest output sequence token count:", outputSequenceTokens.MaxCount, "Sequence:", strings.Join(outputSequenceTokens.LongestSequenceExample, " ")},
		{"Longest input sequence token count:", inputSequenceTokens.MaxCount, "Sequence:", strings.Join(inputSequenceTokens.LongestSequenceExample, " ")},
		{"Longest full sequence token count:", fullSequenceTokens.MaxCount, "Sequence:", strings.Join(fullSequenceTokens.LongestSequenceExample, " ")},
	}
	if err := cursor.WriteValues(values); err != nil {
		return err
	}
	if err := excel.SaveFile(file); err != nil {
		return err
	}
	return nil
}

func getAverage(tokenCount TokenCount) float64 {
	if len(tokenCount.RowsPerTokenCount) == 0 {
		return 0.0
	}

	average := 0.0
	rowCount := 0.0
	for i, count := range tokenCount.RowsPerTokenCount {
		average += float64(count * i)
		rowCount += float64(count)
	}
	return average / rowCount
}

func getInputSequence(method csv.Method) string {
	s := ""
	if utils.ContainsString(method.Modifier, "static") {
		s = "static "
	}
	return fmt.Sprintf("%s%s %s", s, method.ClassName, method.MethodName)
}

func getOutputSequence(parameters []java.Parameter, returnType string) string {
	output := ""
	for i, par := range parameters {
		if i > 0 {
			output += ", "
		}
		output += fmt.Sprintf("%s - %s", par.Type.TypeName, par.Name)
	}
	return output + ". $ " + returnType
}

type TokenCount struct {
	TokenSum               int
	MinCount               int
	MaxCount               int
	RowsPerTokenCount      []int
	LongestSequenceExample []string
}

func (c *TokenCount) Add(tokens []string) {
	tokensCount := len(tokens)
	if c.RowsPerTokenCount == nil {
		c.RowsPerTokenCount = make([]int, tokensCount)
	}
	if len(c.RowsPerTokenCount) <= tokensCount {
		expand := make([]int, (tokensCount+1)-len(c.RowsPerTokenCount))
		c.RowsPerTokenCount = append(c.RowsPerTokenCount, expand...)
	}
	c.RowsPerTokenCount[tokensCount]++

	if c.MaxCount < tokensCount {
		c.MaxCount = tokensCount
		c.LongestSequenceExample = tokens
	}
	if c.MinCount > tokensCount || c.TokenSum == 0 {
		c.MinCount = tokensCount
	}
	c.TokenSum += tokensCount
}

func CreateTokenChart(count TokenCount, title string) excel.Chart {
	series := excel.Series{
		Categories: make([]interface{}, count.MaxCount+1),
		Values:     make([]interface{}, count.MaxCount+1),
	}
	for tokenCount, rowsCount := range count.RowsPerTokenCount {
		series.Categories[tokenCount] = fmt.Sprintf("%d", tokenCount)
		series.Values[tokenCount] = rowsCount
	}

	chart := excel.Chart{
		ChartBase: excel.ChartBase{
			Type: "col",
			Title: &excel.Title{
				Name: title,
			},
			Format: &excel.Format{
				XScale:          1.0,
				YScale:          1.0,
				XOffset:         15,
				YOffset:         10,
				PrintObj:        true,
				LockAspectRatio: false,
				Locked:          false,
			},
			VaryColors: false,
			PlotArea: &excel.PlotArea{
				ShowBubbleSize:  true,
				ShowCatName:     false,
				ShowLeaderLines: false,
				ShowPercent:     true,
				ShowSeriesName:  false,
				ShowVal:         true,
			},
		},
		Series: []excel.Series{series},
	}
	return chart
}
