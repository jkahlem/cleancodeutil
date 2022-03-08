package methodgeneration

type ScoreCalculator interface {
	AddRating(float64)
	Score() float64
}

type F1ScoreCalculator struct{}

func (c *F1ScoreCalculator) AddRating(float64) {}
func (c *F1ScoreCalculator) Score() float64 {
	return 0
}
