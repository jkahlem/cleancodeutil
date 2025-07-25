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
	// Sets the java version which should be considered when parsing the code files.
	JavaVersion int `json:"javaVersion"`
	// If true, the crawler will try to parse incomplete code files (which have parser errors etc.) as good as possible.
	ParseIncomplete bool `json:"parseIncomplete"`
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

func (o *OptionsBuilder) WithJavaVersion(version int) *OptionsBuilder {
	o.options.JavaVersion = version
	return o
}

func (o *OptionsBuilder) WithParseIncomplete(state bool) *OptionsBuilder {
	o.options.ParseIncomplete = state
	return o
}

func (o *OptionsBuilder) Build() Options {
	return o.options
}
