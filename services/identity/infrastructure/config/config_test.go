package config

import (
	"testing"
	"time"
)

func TestDuration(t *testing.T) {
	t.Setenv("TEST_DURATION_KEY", "10m")
	if got := Duration("TEST_DURATION_KEY", time.Minute); got != 10*time.Minute {
		t.Fatalf("Duration = %v, want 10m", got)
	}
	if got := Duration("TEST_DURATION_MISSING", time.Minute); got != time.Minute {
		t.Fatalf("Duration default = %v, want 1m", got)
	}
	t.Setenv("TEST_DURATION_BAD", "not-a-duration")
	if got := Duration("TEST_DURATION_BAD", time.Minute); got != time.Minute {
		t.Fatalf("Duration bad = %v, want 1m", got)
	}
}

func TestDays(t *testing.T) {
	t.Setenv("TEST_DAYS_KEY", "45")
	if got := Days("TEST_DAYS_KEY", 30); got != 45 {
		t.Fatalf("Days = %d, want 45", got)
	}
	if got := Days("TEST_DAYS_MISSING", 30); got != 30 {
		t.Fatalf("Days default = %d, want 30", got)
	}
	t.Setenv("TEST_DAYS_BAD", "abc")
	if got := Days("TEST_DAYS_BAD", 30); got != 30 {
		t.Fatalf("Days bad = %d, want 30", got)
	}
}
