package methodgeneration

import (
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/metrics"
	"strings"

	"github.com/waygo/bleu"
)

type Metric interface {
	Rate(m Method)
	Name() string
	Score() float64
}

type AllZeroRater struct{}

func (r *AllZeroRater) Rate(m Method) {}

func (r *AllZeroRater) Name() string {
	return "All zero"
}

func (r *AllZeroRater) Score() float64 {
	return 0
}

type BleuRater struct {
	score  float64
	count  float64
	config configuration.BleuConfiguration
}

func (r *BleuRater) Rate(m Method) {
	if r.config.Weights == nil {
		r.config.Weights = []float64{0.25, 0.25, 0.25, 0.25}
	}
	r.score += bleu.Compute(r.sentence(m.GeneratedDefinition), []bleu.Sentence{r.sentence(m.ExpectedDefinition)}, r.config.Weights)
	r.count++
}

func (r *BleuRater) sentence(str string) bleu.Sentence {
	return strings.Split(str, " ")
}

func (r *BleuRater) Score() float64 {
	return r.score / r.count
}

func (r *BleuRater) Name() string {
	// TODO: Include options/weights?
	return "Bleu"
}

type RougeType string

const (
	RougeL string = "rouge-l"
	RougeS string = "rouge-s"
	RougeN string = "rouge-n"
)

type RougeRater struct {
	rater     func(m Method) (precision, recall float64)
	precision float64
	recall    float64
	count     float64
}

func RateL(m Method) (precision, recall float64) {
	return 0, 0
}

func NewRougeLRater(config configuration.MetricConfiguration) *RougeRater {
	_, err := config.AsRougeL()
	if err != nil {
		// TODO: remove panic
		panic(err)
	}
	return &RougeRater{
		rater: func(m Method) (precision, recall float64) {
			return metrics.RougeL(m.ExpectedDefinition, []string{m.GeneratedDefinition})
		},
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
			return metrics.RougeN(m.ExpectedDefinition, []string{m.GeneratedDefinition}, c.N)
		},
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
			return metrics.RougeS(m.ExpectedDefinition, []string{m.GeneratedDefinition}, c.SkipN)
		},
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

func (r *RougeRater) Score() float64 {
	return metrics.FScore(r.precision, r.recall, 1)
}
