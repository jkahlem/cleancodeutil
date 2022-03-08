package configuration

type EvaluationConfiguration struct {
	// Defines, how the rating per row should be done, like equality checks or different tools etc.
	RatingTypes []string `json:"ratingTypes"`
	// Defines, how the rating per method should be used to calculate an overall score (per subset)
	ScoreTypes []string `json:"scoreTypes"`
	// Subsets of the evaluation set for which scores should be also calculated (e.g. filter out setter/getter for evaluation and so on)
	Subsets []EvaluationSet `json:"subsets"`
}

type EvaluationSet struct {
	Subsets []EvaluationSet `json:"subsets"`
}
