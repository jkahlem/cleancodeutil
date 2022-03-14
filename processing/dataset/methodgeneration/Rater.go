package methodgeneration

import (
	"strings"

	"github.com/waygo/bleu"
)

type Rater interface {
	Rate(m Method) float64
	Name() string
}

type AllZeroRater struct{}

func (r *AllZeroRater) Rate(m Method) float64 {
	return 0
}

func (r *AllZeroRater) Name() string {
	return "All zero"
}

type BleuRater struct{}

func (r *BleuRater) Rate(m Method) float64 {
	weights := []float64{0.25, 0.25, 0.25, 0.25}
	return bleu.Compute(r.sentence(m.GeneratedDefinition), []bleu.Sentence{r.sentence(m.ExpectedDefinition)}, weights)
}

func (r *BleuRater) sentence(str string) bleu.Sentence {
	return strings.Split(str, " ")
}

func (r *BleuRater) Name() string {
	return "Bleu"
}
