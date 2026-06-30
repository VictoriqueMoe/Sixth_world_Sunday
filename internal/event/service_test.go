package event

import (
	"testing"
	"time"
)

func mustTime(t *testing.T, s string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("parse time %q: %v", s, err)
	}
	return parsed.UTC()
}

func assertOccurrences(t *testing.T, got []time.Time, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d occurrences, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if g := got[i].Format(time.RFC3339); g != want[i] {
			t.Errorf("occurrence %d = %s, want %s", i, g, want[i])
		}
	}
}

func TestNextOccurrencesNone(t *testing.T) {
	start := mustTime(t, "2026-07-05T20:00:00Z")

	future := nextOccurrences(start, "none", mustTime(t, "2026-07-01T00:00:00Z"), 5)
	assertOccurrences(t, future, []string{"2026-07-05T20:00:00Z"})

	past := nextOccurrences(start, "none", mustTime(t, "2026-07-06T00:00:00Z"), 5)
	if past != nil {
		t.Errorf("expected nil for past one-off, got %v", past)
	}
}

func TestNextOccurrencesWeekly(t *testing.T) {
	start := mustTime(t, "2026-07-05T20:00:00Z")
	now := mustTime(t, "2026-07-06T00:00:00Z")

	got := nextOccurrences(start, "weekly", now, 3)
	assertOccurrences(t, got, []string{
		"2026-07-12T20:00:00Z",
		"2026-07-19T20:00:00Z",
		"2026-07-26T20:00:00Z",
	})
}

func TestNextOccurrencesBiweekly(t *testing.T) {
	start := mustTime(t, "2026-07-05T20:00:00Z")
	now := mustTime(t, "2026-07-06T00:00:00Z")

	got := nextOccurrences(start, "biweekly", now, 3)
	assertOccurrences(t, got, []string{
		"2026-07-19T20:00:00Z",
		"2026-08-02T20:00:00Z",
		"2026-08-16T20:00:00Z",
	})
}

func TestNextOccurrencesMonthlyNthWeekday(t *testing.T) {
	// 2026-07-05 is the first Sunday of July.
	start := mustTime(t, "2026-07-05T20:00:00Z")
	now := mustTime(t, "2026-07-06T00:00:00Z")

	got := nextOccurrences(start, "monthly", now, 3)
	assertOccurrences(t, got, []string{
		"2026-08-02T20:00:00Z", // first Sunday of August
		"2026-09-06T20:00:00Z", // first Sunday of September
		"2026-10-04T20:00:00Z", // first Sunday of October
	})
}

func TestNextOccurrencesAnnually(t *testing.T) {
	start := mustTime(t, "2026-07-05T20:00:00Z")
	now := mustTime(t, "2026-07-06T00:00:00Z")

	got := nextOccurrences(start, "annually", now, 2)
	assertOccurrences(t, got, []string{
		"2027-07-05T20:00:00Z",
		"2028-07-05T20:00:00Z",
	})
}

func TestNextOccurrencesRollsForwardFromPastStart(t *testing.T) {
	start := mustTime(t, "2026-01-04T20:00:00Z") // a Sunday far in the past
	now := mustTime(t, "2026-07-06T00:00:00Z")

	got := nextOccurrences(start, "weekly", now, 1)
	if len(got) != 1 {
		t.Fatalf("expected 1 occurrence, got %d", len(got))
	}
	if got[0].Before(now) {
		t.Errorf("first occurrence %s is before now %s", got[0], now)
	}
	if got[0].Weekday() != time.Sunday {
		t.Errorf("expected Sunday, got %s", got[0].Weekday())
	}
}
