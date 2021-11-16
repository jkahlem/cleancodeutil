package charts

import (
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// A bar chart builder wrapping the bar chart building functionalities of go-echarts
type BarChartBuilder struct {
	title     string
	subTitle  string
	dataTitle string
	data      map[string]int
}

func NewBarChartBuilder() *BarChartBuilder {
	return &BarChartBuilder{}
}

// Sets the title of the bar chart
func (builder *BarChartBuilder) WithTitle(title, subTitle string) *BarChartBuilder {
	builder.title = title
	builder.subTitle = subTitle
	return builder
}

// Sets the title of the data of the bar chart. This will for example appear in the tooltip when hovering over a bar.
func (builder *BarChartBuilder) WithDataTitle(dataTitle string) *BarChartBuilder {
	builder.dataTitle = dataTitle
	return builder
}

// Adds data to the bar chart.
func (builder *BarChartBuilder) AddData(label string, value int) *BarChartBuilder {
	if builder.data == nil {
		builder.data = make(map[string]int)
	}
	builder.data[label] = value
	return builder
}

// Builds the bar chart.
func (builder *BarChartBuilder) Build() *charts.Bar {
	chart := charts.NewBar()
	chart.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    builder.title,
		Subtitle: builder.subTitle,
	}))
	xAxis, data := builder.getSortedDataAndAxisSlices()
	chart.SetXAxis(xAxis)
	chart.AddSeries(builder.dataTitle, data)
	return chart
}

// Returns data sorted together with the axis headings.
func (builder *BarChartBuilder) getSortedDataAndAxisSlices() (axis []string, data []opts.BarData) {
	axis = make([]string, 0, len(builder.data))
	data = make([]opts.BarData, 0, len(builder.data))
	for i := 0; i < cap(data); i++ {
		maxValue, maxLabel := -1, ""
		for label, value := range builder.data {
			if value > maxValue {
				maxValue = value
				maxLabel = label
			}
		}
		delete(builder.data, maxLabel)
		axis = append(axis, maxLabel)
		data = append(data, opts.BarData{
			Value:   maxValue,
			Tooltip: &opts.Tooltip{Show: true},
		})
	}
	return
}
