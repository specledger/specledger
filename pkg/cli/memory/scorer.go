package memory

// CalculateComposite computes the composite score as the average of three axes.
func CalculateComposite(s Score) float64 {
	return (s.Recurrence + s.Impact + s.Specificity) / 3.0
}

// ShouldPromote returns true if the composite score meets or exceeds the promotion threshold.
func ShouldPromote(s Score) bool {
	return CalculateComposite(s) >= PromotionThreshold
}

// NewScore creates a Score with a calculated composite.
func NewScore(recurrence, impact, specificity float64) Score {
	s := Score{
		Recurrence:  recurrence,
		Impact:      impact,
		Specificity: specificity,
	}
	s.Composite = CalculateComposite(s)
	return s
}
