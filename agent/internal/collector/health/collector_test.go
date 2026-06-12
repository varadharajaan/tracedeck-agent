package health

import (
	"testing"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
)

func TestHealthScoreBounds(t *testing.T) {
	t.Parallel()

	if got := healthScore(0, 0, 0); got != 100 {
		t.Fatalf("expected perfect score 100, got %d", got)
	}
	if got := healthScore(1000, 1000, 1000); got != 0 {
		t.Fatalf("expected score floor 0, got %d", got)
	}
}

func TestHealthStatus(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		score int
		want  string
	}{
		{score: 90, want: constants.HealthStatusHealthy},
		{score: 70, want: constants.HealthStatusWatch},
		{score: 40, want: constants.HealthStatusAttention},
	} {
		if got := healthStatus(tc.score); got != tc.want {
			t.Fatalf("healthStatus(%d) = %q, want %q", tc.score, got, tc.want)
		}
	}
}
