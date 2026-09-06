package store

import (
	"testing"
)

func TestEvaluateDeterminism(t *testing.T) {
	s := New()
	s.flags["f"] = Flag{Key: "f", Enabled: true, RolloutPercent: 50}

	first, err := s.Evaluate("f", "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 0; i < 100; i++ {
		got, err := s.Evaluate("f", "alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != first {
			t.Fatalf("non-deterministic result: %v then %v", first, got)
		}
	}
}

func TestEvaluateDisabledIsFalse(t *testing.T) {
	s := New()
	s.flags["f"] = Flag{Key: "f", Enabled: false, RolloutPercent: 100}

	got, err := s.Evaluate("f", "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Fatalf("expected false for disabled flag, got true")
	}
}

func TestEvaluateRollout100IsTrue(t *testing.T) {
	s := New()
	s.flags["f"] = Flag{Key: "f", Enabled: true, RolloutPercent: 100}

	got, err := s.Evaluate("f", "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Fatalf("expected true for rollout_percent 100, got false")
	}
}

func TestEvaluateRollout0IsFalse(t *testing.T) {
	s := New()
	s.flags["f"] = Flag{Key: "f", Enabled: true, RolloutPercent: 0}

	got, err := s.Evaluate("f", "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Fatalf("expected false for rollout_percent 0, got true")
	}
}

func TestEvaluateRespectsIntermediatePercent(t *testing.T) {
	s := New()
	s.flags["f"] = Flag{Key: "f", Enabled: true, RolloutPercent: 50}

	seen := map[bool]bool{}
	for _, user := range []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"} {
		got, err := s.Evaluate("f", user)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		seen[got] = true
	}
	if len(seen) != 2 {
		t.Fatalf("expected a mix of true and false for rollout_percent 50, got %v", seen)
	}
}

func TestEvaluateUnknownKey(t *testing.T) {
	s := New()

	_, err := s.Evaluate("missing", "alice")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
