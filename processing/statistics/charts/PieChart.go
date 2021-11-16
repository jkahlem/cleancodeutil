package charts

import (
	"fmt"
	"returntypes-langserver/common/utils"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

type PieData struct {
	Label string
	Value int
	Color string
}

// A pie chart builder wrapping the pie chart building functionalities of go-echarts
type PieChartBuilder struct {
	title               string
	subTitle            string
	data                map[string]PieData
	dataTitle           string
	hideTooltip         bool
	percentageOnly      bool
	threshold           float64
	thresholdGroupColor string
}

func NewPieChartBuilder() *PieChartBuilder {
	return &PieChartBuilder{}
}

// Sets the title of the pie chart
func (builder *PieChartBuilder) WithTitle(title, subTitle string) *PieChartBuilder {
	builder.title = title
	builder.subTitle = subTitle
	return builder
}

// Sets the title of the data of the pie chart. This will for example appear in the tooltip when hovering the chart.
func (builder *PieChartBuilder) WithDataTitle(title string) *PieChartBuilder {
	builder.dataTitle = title
	return builder
}

// Hides the tooltip if set to true.
func (builder *PieChartBuilder) WithHideTooltip(state bool) *PieChartBuilder {
	builder.hideTooltip = state
	return builder
}

// If set to true, the labels of each part of the pie chart will only show the percentage information and not the real value.
func (builder *PieChartBuilder) WithPercentageOnly(state bool) *PieChartBuilder {
	builder.percentageOnly = state
	return builder
}

// Sets a threshold for the pie chart. If the amount of the data goes under this threshold, it will be grouped in one group.
func (builder *PieChartBuilder) WithThreshold(threshold float64, thresholdGroupColor string) *PieChartBuilder {
	builder.threshold = threshold
	builder.thresholdGroupColor = thresholdGroupColor
	return builder
}

// Adds data to the pie chart.
func (builder *PieChartBuilder) AddData(label string, value int) *PieChartBuilder {
	if builder.data == nil {
		builder.data = make(map[string]PieData)
	}
	builder.data[builder.generateUuid()] = PieData{
		Label: label,
		Value: value,
	}
	return builder
}

// Adds data with specific colors to the pie chart.
func (builder *PieChartBuilder) AddDataWithColor(label string, value int, color string) *PieChartBuilder {
	if builder.data == nil {
		builder.data = make(map[string]PieData)
	}
	builder.data[builder.generateUuid()] = PieData{
		Label: label,
		Value: value,
		Color: color,
	}
	return builder
}

func (builder *PieChartBuilder) generateUuid() string {
	return utils.NewUuid()
}

// Builds the pie chart.
func (builder *PieChartBuilder) Build() *charts.Pie {
	chart := charts.NewPie()
	chart.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    builder.title,
		Subtitle: builder.subTitle,
	}))
	chart.AddSeries(builder.dataTitle, builder.getSortedSeries()).
		SetSeriesOptions(charts.WithLabelOpts(opts.Label{
			Show:      true,
			Formatter: builder.formatter(),
		}))
	return chart
}

// Sorts the data in descending order by value
func (builder *PieChartBuilder) getSortedSeries() []opts.PieData {
	under, over := builder.splitDataByThreshold()
	data := make([]opts.PieData, 0, len(over))
	for i := 0; i < cap(data); i++ {
		maxValue, maxIndex := PieData{Value: -1}, -1
		for j, value := range over {
			if value.Value == -1 {
				continue
			} else if maxValue.Value < value.Value {
				maxValue = value
				maxIndex = j
			}
		}
		over[maxIndex].Value = -1
		data = append(data, builder.mapPieData(maxValue))
	}
	if len(under) == 1 {
		// no need to summarize/group one single value, so just add it to the data
		// (as it's the only one under the threshold, it's already the lowest value)
		data = append(data, builder.mapPieData(under[0]))
	} else if len(under) > 1 {
		// Otherwise, summarize the data under the threshold
		summarized := builder.summarizePieData(under, fmt.Sprintf("Other (%d)", len(under)), builder.thresholdGroupColor)
		data = append(data, builder.mapPieData(summarized))
	}
	return data
}

// Splits the builder data by the configured threshold. If the threshold is not set or there is only one data under the threshold
// then the function will return nil for under
func (builder *PieChartBuilder) splitDataByThreshold() (under, over []PieData) {
	if builder.threshold > 0 {
		sum := builder.calculateSum()
		if sum > 0 {
			under, over = make([]PieData, 0, len(builder.data)), make([]PieData, 0, len(builder.data))
			for _, data := range builder.data {
				fraction := float64(data.Value) / float64(sum)
				if fraction < builder.threshold {
					under = append(under, data)
				} else {
					over = append(over, data)
				}
			}
			return under, over
		}
	}
	return nil, builder.mapDataToSlice()
}

// Calculates the sum of the values of the data
func (builder *PieChartBuilder) calculateSum() int {
	sum := 0
	for _, data := range builder.data {
		sum += data.Value
	}
	return sum
}

// Maps the data to a slice
func (builder *PieChartBuilder) mapDataToSlice() []PieData {
	destination := make([]PieData, 0, len(builder.data))
	for _, data := range builder.data {
		destination = append(destination, data)
	}
	return destination
}

// Summarizes the given data to a PieData object with the given label
func (builder *PieChartBuilder) summarizePieData(data []PieData, label, color string) PieData {
	out := PieData{
		Label: label,
		Value: 0,
		Color: color,
	}
	for _, v := range data {
		out.Value += v.Value
	}
	return out
}

// Maps the pie data to the go-echarts pie data
func (builder *PieChartBuilder) mapPieData(data PieData) opts.PieData {
	pieData := opts.PieData{
		Name:    data.Label,
		Value:   data.Value,
		Tooltip: &opts.Tooltip{Show: !builder.hideTooltip},
	}
	if data.Color != "" {
		pieData.ItemStyle = &opts.ItemStyle{
			Color: data.Color,
		}
	}
	return pieData
}

// Formats the label of each value in the pie chart
func (builder *PieChartBuilder) formatter() string {
	if builder.percentageOnly {
		return opts.FuncOpts(`function(data) { return data.name + ': ' + data.percent + '%'; }`)
	}
	return opts.FuncOpts(`function(data) { return data.name + ': ' + data.value + ' (' + data.percent + '%)'; }`)
}
