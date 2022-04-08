package excel

type Title struct {
	Name string `json:"name"`
}

type Format struct {
	XScale          float64 `json:"x_scale"`
	YScale          float64 `json:"y_scale"`
	XOffset         int     `json:"x_offset"`
	YOffset         int     `json:"y_offset"`
	PrintObj        bool    `json:"print_obj"`
	LockAspectRatio bool    `json:"lock_aspect_ratio"`
	Locked          bool    `json:"locked"`
}

type Legend struct {
	Position      string `json:"position"`
	ShowLegendKey bool   `json:"show_legend_key"`
}

type PlotArea struct {
	ShowBubbleSize  bool `json:"show_bubble_size"`
	ShowCatName     bool `json:"show_cat_name"`
	ShowLeaderLines bool `json:"show_leader_lines"`
	ShowPercent     bool `json:"show_percent"`
	ShowSeriesName  bool `json:"show_series_name"`
	ShowVal         bool `json:"show_val"`
}

type SeriesRaw struct {
	Name       string `json:"name"`
	Categories string `json:"categories"`
	Values     string `json:"values"`
}

type Series struct {
	Name       string
	Categories []interface{}
	Values     []interface{}
}

type ChartBase struct {
	Type       string    `json:"type"`
	Title      *Title    `json:"title,omitempty"`
	Format     *Format   `json:"format,omitempty"`
	Legend     *Legend   `json:"legend,omitempty"`
	VaryColors bool      `json:"vary_colors"`
	PlotArea   *PlotArea `json:"plotarea,omitempty"`
}

type Chart struct {
	ChartBase
	Series []Series
}

type ChartRaw struct {
	ChartBase
	Series []SeriesRaw `json:"series,omitempty"`
}
