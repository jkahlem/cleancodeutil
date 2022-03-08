package methodgeneration

import (
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/processing/dataset/base"
)

type Evaluator struct{}

type Method struct {
	Name                string
	ExpectedDefinition  string
	GeneratedDefinition string
	Ratings             []Rating
}

type Rating struct {
	Type   string
	Rating float64
}

func NewEvaluator() base.Evaluator {
	return &Evaluator{}
}

func (e *Evaluator) Evaluate() errors.Error {
	set := e.loadEvaluationSet()
	methods := e.generateMethodDefinitions(set)
	evalset := e.getEvaluationSetConfig()

	for _, m := range methods {
		e.rateMethod(&m)
		evalset.AddMethod(m)
	}

	return nil
}

func (e *Evaluator) loadEvaluationSet() [][]string {
	return nil
}

func (e *Evaluator) generateMethodDefinitions(evaluationSet [][]string) []Method {
	return nil
}

func (e *Evaluator) getEvaluationSetConfig() *EvaluationSet {
	set := e.buildEvaluationSet(configuration.EvaluationSet{
		Subsets: configuration.EvaluationSubsets(),
	})
	return &set
}

func (e *Evaluator) buildEvaluationSet(setConfiguration configuration.EvaluationSet) EvaluationSet {
	set := EvaluationSet{
		Subsets: make([]EvaluationSet, 0),
	}

	for _, subset := range setConfiguration.Subsets {
		set.Subsets = append(set.Subsets, e.buildEvaluationSet(subset))
	}
	return set
}

func (e *Evaluator) rateMethod(m *Method) {
	rater := e.getAvailableRater()
	for _, r := range rater {
		rate := r.Rate(*m)
		m.Ratings = append(m.Ratings, Rating{
			Type:   r.Name(),
			Rating: rate,
		})
	}
}

func (e *Evaluator) getAvailableRater() []Rater {
	// TODO: Use configuration.EvaluationRatingTypes() to determine which rater to add
	return []Rater{&AllZeroRater{}}
}

type EvaluationSet struct {
	Subsets []EvaluationSet
	// String -> rating type, ScoreCalculator -> holds score information for that specific rating type
	OverallScore map[string][]ScoreCalculator
}

func (e *EvaluationSet) AddMethod(m Method) {
	if !e.IsMethodAccepted(m) {
		return
	}
	// TOOD:
	// - Add to output?
	e.addRatingsToScore(m.Ratings)
	for i := range e.Subsets {
		e.Subsets[i].AddMethod(m)
	}
}

func (e *EvaluationSet) addRatingsToScore(ratings []Rating) {
	if e.OverallScore == nil {
		e.OverallScore = make(map[string][]ScoreCalculator)
	}
	for _, r := range ratings {
		if _, ok := e.OverallScore[r.Type]; !ok {
			e.initScoreCalculator(r.Type)
		}
		for _, calculator := range e.OverallScore[r.Type] {
			calculator.AddRating(r.Rating)
		}
	}
}

func (e *EvaluationSet) initScoreCalculator(ratingType string) {
	// TODO: Use configuration.EvaluationRatingTypes() to determine which score calculator to add
	e.OverallScore[ratingType] = make([]ScoreCalculator, 0, 5)
	e.OverallScore[ratingType] = append(e.OverallScore[ratingType], &F1ScoreCalculator{})
}

func (e *EvaluationSet) IsMethodAccepted(m Method) bool {
	// TODO
	return true
}
