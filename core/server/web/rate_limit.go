package web

import "time"

// pruneAttemptsAfter keeps timestamps strictly after cutoff (in-place slice reuse).
func pruneAttemptsAfter(attempts []time.Time, cutoff time.Time) []time.Time {
	out := attempts[:0]
	for _, attemptTime := range attempts {
		if attemptTime.After(cutoff) {
			out = append(out, attemptTime)
		}
	}
	return out
}

// pruneAttemptsWithin keeps timestamps within window ending at now.
func pruneAttemptsWithin(attempts []time.Time, now time.Time, window time.Duration) []time.Time {
	cutoff := now.Add(-window)
	return pruneAttemptsAfter(attempts, cutoff)
}
