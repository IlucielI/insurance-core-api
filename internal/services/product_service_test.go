package services

import "testing"

func TestCalculateTermFactor(t *testing.T) {
	if got := calculateTermFactor(5, 10); got != 1.05 {
		t.Fatalf("calculateTermFactor() = %v, want 1.05", got)
	}
}
