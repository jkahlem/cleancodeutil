package methodgeneration

import (
	"fmt"
	"regexp"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/excel"
	"returntypes-langserver/common/metrics"
	"returntypes-langserver/common/utils"
	"returntypes-langserver/services/predictor"
	"strings"

	"github.com/waygo/bleu"
)

type Metric interface {
	Rate(m Method)
	Name() string
	Result() [][]interface{}
}

type AllZeroRater struct{}

func (r *AllZeroRater) Rate(m Method) {}

func (r *AllZeroRater) Name() string {
	return "All zero"
}

func (r *AllZeroRater) Result() [][]interface{} {
	return [][]interface{}{{"Score", "0"}}
}

type BleuRater struct {
	score               float64
	count               float64
	config              configuration.BleuConfiguration
	scoresPerTokenCount map[int]float64
	corpusCandidate     []string
	corpusReference     []string
}

func (r *BleuRater) Rate(m Method) {
	if r.config.Weights == nil {
		r.config.Weights = []float64{0.25, 0.25, 0.25, 0.25}
	}
	r.corpusCandidate = append(r.corpusCandidate, m.GeneratedDefinition.Tokens()...)
	r.corpusReference = append(r.corpusReference, m.ExpectedDefinition.Tokens()...)
	r.score += bleu.Smooth(r.sentence(m.GeneratedDefinition), []bleu.Sentence{r.sentence(m.ExpectedDefinition)}, r.config.Weights)
	r.count++
}

func (r *BleuRater) sentence(sentence *metrics.Sentence) bleu.Sentence {
	return sentence.Tokens()
}

func (r *BleuRater) Result() [][]interface{} {
	corpusBleu := bleu.Compute(r.corpusCandidate, []bleu.Sentence{r.corpusReference}, r.config.Weights)
	return [][]interface{}{{"Sentence score average", r.score / r.count},
		{"Corpus Score (without smooth)", corpusBleu}}
}

func (r *BleuRater) Name() string {
	// TODO: Include options/weights?
	return fmt.Sprintf("Smoothed Bleu (Weights: %s)", r.weights())
}

func (r *BleuRater) weights() string {
	output := ""
	for i, weight := range r.config.Weights {
		if i > 0 {
			output += ", "
		}
		output += fmt.Sprintf("%f", weight)
	}
	return output
}

type RougeType string

const (
	RougeL string = "rouge-l"
	RougeS string = "rouge-s"
	RougeN string = "rouge-n"
)

type RougeRater struct {
	rater     func(m Method) (precision, recall float64)
	measure   configuration.Measure
	precision float64
	recall    float64
	count     float64
}

func NewRougeLRater(config configuration.MetricConfiguration) *RougeRater {
	c, err := config.AsRougeL()
	if err != nil {
		// TODO: remove panic
		panic(err)
	}
	return &RougeRater{
		rater: func(m Method) (precision, recall float64) {
			return metrics.RougeL(m.ExpectedDefinition, []*metrics.Sentence{m.GeneratedDefinition})
		},
		measure: c.Measure,
	}
}

func NewRougeNRater(config configuration.MetricConfiguration) *RougeRater {
	c, err := config.AsRougeN()
	if err != nil {
		// TODO: remove panic
		panic(err)
	}
	return &RougeRater{
		rater: func(m Method) (precision, recall float64) {
			return metrics.RougeN(m.ExpectedDefinition, []*metrics.Sentence{m.GeneratedDefinition}, c.N)
		},
		measure: c.Measure,
	}
}

func NewRougeSRater(config configuration.MetricConfiguration) *RougeRater {
	c, err := config.AsRougeS()
	if err != nil {
		// TODO: remove panic
		panic(err)
	}
	return &RougeRater{
		rater: func(m Method) (precision, recall float64) {
			return metrics.RougeS(m.ExpectedDefinition, []*metrics.Sentence{m.GeneratedDefinition}, c.SkipN)
		},
		measure: c.Measure,
	}
}

func (r *RougeRater) Rate(m Method) {
	precision, recall := r.rater(m)
	r.precision += precision
	r.recall += recall
	r.count++
}

func (r *RougeRater) sentence(str string) []string {
	return strings.Split(str, " ")
}

func (r *RougeRater) Name() string {
	// TODO: Include options/weights?
	return "Rouge"
}

func (r *RougeRater) score() float64 {
	switch r.measure.Type() {
	case configuration.FScore:
		if fscore, err := r.measure.AsFScore(); err != nil {
			// TODO: remove panic
			panic(err)
		} else {
			return metrics.FScore(r.precision/r.count, r.recall/r.count, fscore.Beta)
		}
	default:
		return metrics.FScore(r.precision/r.count, r.recall/r.count, 1)
	}
}

func (r *RougeRater) Result() [][]interface{} {
	return [][]interface{}{{"Score", fmt.Sprintf("%f", r.score())}}
}

type IdealRater struct{}

func (r *IdealRater) Rate(m Method) {}
func (r *IdealRater) Name() string {
	return "Ideal"
}
func (r *IdealRater) Result() [][]interface{} {
	return nil
}

type TokenCounter struct {
	expectedTokenCount  TokenCount
	generatedTokenCount TokenCount
	rowsCount           int
}

func (r *TokenCounter) Rate(m Method) {
	r.expectedTokenCount.Add(m.ExpectedDefinition.Ngram(1))
	r.generatedTokenCount.Add(m.GeneratedDefinition.Ngram(1))
	r.rowsCount++
}

func (r *TokenCounter) Name() string {
	return "Parameter counter"
}

func (r *TokenCounter) Result() [][]interface{} {
	expectedDefinitionsResult := r.resultFor(r.expectedTokenCount)
	generatedDefinitionsResult := r.resultFor(r.generatedTokenCount)
	result := make([][]interface{}, 0, len(expectedDefinitionsResult)+len(generatedDefinitionsResult)+2)

	result = append(result, []interface{}{excel.Markdown("**Expected Definitions**")})
	result = append(result, expectedDefinitionsResult...)

	result = append(result, []interface{}{excel.Markdown("**Generated Definitions**")})
	result = append(result, expectedDefinitionsResult...)
	return result
}

func (r *TokenCounter) resultFor(count TokenCount) [][]interface{} {
	outputs := [][]interface{}{
		{"Overall number of tokens", fmt.Sprintf("%d in %d sequences", count.TokenSum, r.rowsCount)},
		{"Minimum of tokens in one output sequence", count.MinCount},
		{"Maximum of tokens in one output sequence", count.MaxCount},
		{"Average token count per sequence", float64(count.TokenSum) / float64(r.rowsCount)},
		{},
	}
	outputs = append(outputs, r.tokenMap(count))
	return outputs
}

func (r *TokenCounter) tokenMap(count TokenCount) []interface{} {
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
				Name: "Tokens per parameter list",
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
	return []interface{}{chart}
}

type TokenCount struct {
	TokenSum          int
	MinCount          int
	MaxCount          int
	RowsPerTokenCount []int
}

func (c *TokenCount) Add(tokens metrics.Ngram) {
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
	}
	if c.MinCount > tokensCount || c.TokenSum == 0 {
		c.MinCount = tokensCount
	}
	c.TokenSum += tokensCount
}

type ExactRater struct {
	matches float64
	count   float64
}

func (r *ExactRater) Rate(m Method) {
	if r.isMatching(m.GeneratedDefinition.Tokens(), m.ExpectedDefinition.Tokens()) {
		r.matches++
	}
	r.count++
}

func (r *ExactRater) isMatching(generated, expected []string) bool {
	if len(generated) != len(expected) {
		return false
	}
	for i := range generated {
		if generated[i] != expected[i] {
			return false
		}
	}
	return true
}

func (r *ExactRater) Result() [][]interface{} {
	return [][]interface{}{{"Average", r.matches / r.count},
		{"Matches", r.matches},
		{"Overall count", r.count}}
}

func (r *ExactRater) Name() string {
	// TODO: Include options/weights?
	return fmt.Sprintf("Exact matches")
}

type CompilabilityRater struct {
	validCount float64
	count      float64
}

var TokenMatcher = regexp.MustCompile("^[a-zA-Z_][a-zA-Z0-9_]*$")

func (r *CompilabilityRater) Rate(m Method) {
	if !r.hasCompileErrors(m.Method) {
		r.validCount++
	}
	r.count++
}

func (r *CompilabilityRater) hasCompileErrors(method predictor.Method) bool {
	// Checks if:
	// - tokens (parameter names, types and return types) consist of valid characters ([a-zA-Z_][a-zA-Z0-9_]*)
	//   - for example, parts of the tokens like [arr] and so on should not be present.
	//   - this checks also, if the tokens are empty
	// - parameter names do not overlap with other parameter names
	parameterNames := make(utils.StringSet)
	for _, par := range method.Values.Parameters {
		concatenatedName := ConcatByLowerCamelCase(strings.Split(par.Name, " "))
		if parameterNames.Has(concatenatedName) || !TokenMatcher.Match([]byte(concatenatedName)) {
			return true
		} else {
			parameterNames.Put(concatenatedName)
		}

		if !r.isValidTypeIdentifier(par.Type) {
			return true
		}
	}

	return !r.isValidTypeIdentifier(method.Values.ReturnType)
}

func (r *CompilabilityRater) isValidTypeIdentifier(typeIdentifier string) bool {
	concatenated := ConcatByUpperCamelCase(strings.Split(typeIdentifier, " "))
	return TokenMatcher.Match([]byte(concatenated))
}

func (r *CompilabilityRater) Result() [][]interface{} {
	return [][]interface{}{{"Average", r.validCount / r.count},
		{"Valid methods", r.validCount},
		{"Overall count", r.count}}
}

func (r *CompilabilityRater) Name() string {
	return fmt.Sprintf("Compilability rate")
}
