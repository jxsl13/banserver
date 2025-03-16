package parser

import (
	"math"
	"strconv"
	"testing"
)

func TestParseInt(t *testing.T) {
	minInt := strconv.Itoa(math.MinInt64)
	maxInt := strconv.Itoa(math.MaxInt64)

	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"min int", minInt, math.MinInt64},
		{"max int", maxInt, math.MaxInt64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := parseInt(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if actual != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, actual)
			}
		})
	}
}
