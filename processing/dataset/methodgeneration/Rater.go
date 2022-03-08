package methodgeneration

type Rater interface {
	Rate(m Method) float64
	Name() string
}

type AllZeroRater struct{}

func (r *AllZeroRater) Rate(m Method) float64 {
	return 0
}

func (r *AllZeroRater) Name() string {
	return "All zero rater"
}
