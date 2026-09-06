package store

import (
	"fmt"
	"sync"
	"testing"
)

func TestCreateAndGet(t *testing.T) {
	s := New()
	f := Flag{Key: "a", Enabled: true, Description: "desc", RolloutPercent: 50}
	if err := s.Create(f); err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, ok := s.Get("a")
	if !ok {
		t.Fatal("Get: expected flag to exist")
	}
	if got != f {
		t.Fatalf("Get: got %+v want %+v", got, f)
	}
}

func TestCreateDuplicate(t *testing.T) {
	s := New()
	if err := s.Create(Flag{Key: "a"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := s.Create(Flag{Key: "a"}); err != ErrDuplicate {
		t.Fatalf("Create duplicate: got %v want %v", err, ErrDuplicate)
	}
}

func TestListSorted(t *testing.T) {
	s := New()
	for _, key := range []string{"b", "a", "c"} {
		if err := s.Create(Flag{Key: key}); err != nil {
			t.Fatalf("Create %q: %v", key, err)
		}
	}
	list := s.List()
	if len(list) != 3 {
		t.Fatalf("List len: got %d want 3", len(list))
	}
	if list[0].Key != "a" || list[1].Key != "b" || list[2].Key != "c" {
		t.Fatalf("List order: got %v", list)
	}
}

func TestListEmpty(t *testing.T) {
	s := New()
	list := s.List()
	if len(list) != 0 {
		t.Fatalf("List empty: got %d want 0", len(list))
	}
	if list == nil {
		t.Fatal("List empty: expected non-nil slice so it encodes to []")
	}
}

func TestGetNotFound(t *testing.T) {
	s := New()
	if _, ok := s.Get("missing"); ok {
		t.Fatal("Get missing: expected not ok")
	}
}

func TestUpdate(t *testing.T) {
	s := New()
	if err := s.Create(Flag{Key: "a", Enabled: false}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	updated, ok := s.Update("a", Flag{Enabled: true, RolloutPercent: 10})
	if !ok {
		t.Fatal("Update: expected ok")
	}
	if updated.Key != "a" || !updated.Enabled || updated.RolloutPercent != 10 {
		t.Fatalf("Update: got %+v", updated)
	}
}

func TestUpdateNotFound(t *testing.T) {
	s := New()
	if _, ok := s.Update("missing", Flag{}); ok {
		t.Fatal("Update missing: expected not ok")
	}
}

func TestDelete(t *testing.T) {
	s := New()
	if err := s.Create(Flag{Key: "a"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if !s.Delete("a") {
		t.Fatal("Delete: expected true")
	}
	if _, ok := s.Get("a"); ok {
		t.Fatal("Delete: flag still exists")
	}
}

func TestDeleteNotFound(t *testing.T) {
	s := New()
	if s.Delete("missing") {
		t.Fatal("Delete missing: expected false")
	}
}

func TestConcurrentAccess(t *testing.T) {
	s := New()
	const goroutines = 50
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("flag-%d", i%10)
			if err := s.Create(Flag{Key: key}); err != nil && err != ErrDuplicate {
				t.Errorf("Create: %v", err)
			}
			s.Get(key)
			s.List()
			s.Update(key, Flag{Key: key, Enabled: true})
			s.Delete(key)
		}(i)
	}
	wg.Wait()
}
