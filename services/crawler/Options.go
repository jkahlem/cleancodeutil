package crawler

// Defines possible options which can be used for the crawling process.
type Options struct {
	// Save ranges of methods definitions in XML output
	UseRanges bool `json:"useRanges"`
	// Paths should be saved as absolute paths
	UseAbsolutePaths bool `json:"useAbsolutePaths"`
	// If true, the crawler will not stop crawling if a file could not be parsed
	Forced bool `json:"forced"`
	// If true, the crawler will not send log notifications
	Silent bool `json:"silent"`
}

type OptionsBuilder struct {
	options Options
}

func NewOptions() *OptionsBuilder {
	return &OptionsBuilder{}
}

func (o *OptionsBuilder) WithRanges(state bool) *OptionsBuilder {
	o.options.UseRanges = state
	return o
}

func (o *OptionsBuilder) WithAbsolutePaths(state bool) *OptionsBuilder {
	o.options.UseAbsolutePaths = state
	return o
}

func (o *OptionsBuilder) Forced(state bool) *OptionsBuilder {
	o.options.Forced = state
	return o
}

func (o *OptionsBuilder) Silent(state bool) *OptionsBuilder {
	o.options.Silent = state
	return o
}

func (o *OptionsBuilder) Build() Options {
	return o.options
}
