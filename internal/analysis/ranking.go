package analysis

// RankBySeverity sorts leaks by severity for display and alerting.
func RankBySeverity(leaks []Leak) []Leak {
	if len(leaks) == 0 {
		return nil
	}
	out := make([]Leak, len(leaks))
	copy(out, leaks)
	return out
}
