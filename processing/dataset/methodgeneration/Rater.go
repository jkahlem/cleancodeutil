package methodgeneration

import (
	"returntypes-langserver/common/metrics"
	"strings"

	"github.com/waygo/bleu"
)

type Rater interface {
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
	score float64
	count float64
}

func (r *BleuRater) Rate(m Method) {
	weights := []float64{0.25, 0.25, 0.25, 0.25}
	r.score += bleu.Compute(r.sentence(m.GeneratedDefinition), []bleu.Sentence{r.sentence(m.ExpectedDefinition)}, weights)
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
	RougeL RougeType = "rouge-l"
	RougeS RougeType = "rouge-s"
	RougeN RougeType = "rouge-n"
)

type RougeRater struct {
	Type      RougeType
	precision float64
	recall    float64
	count     float64
}

func (r *RougeRater) Rate(m Method) {
	var precision, recall float64
	// TODO: Add more specific metric options
	switch r.Type {
	case RougeL:
		precision, recall = metrics.RougeL(m.ExpectedDefinition, []string{m.GeneratedDefinition})
	case RougeN:
		precision, recall = metrics.RougeN(m.ExpectedDefinition, []string{m.GeneratedDefinition}, 1)
	case RougeS:
		precision, recall = metrics.RougeS(m.ExpectedDefinition, []string{m.GeneratedDefinition}, 1)
	}
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
