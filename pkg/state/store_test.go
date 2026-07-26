package state

import (
	"sync"
	"testing"
)

func TestGetSet(t *testing.T) {
	s := New()
	s.Set("user.name", "Alice")

	val, ok := s.Get("user.name")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if val != "Alice" {
		t.Errorf("got %v, want Alice", val)
	}
}

func TestGetMissing(t *testing.T) {
	s := New()
	_, ok := s.Get("missing")
	if ok {
		t.Error("expected key to be missing")
	}
}

func TestNestedSet(t *testing.T) {
	s := New()
	s.Set("a.b.c", 42)

	val, ok := s.Get("a.b.c")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if val != 42 {
		t.Errorf("got %v, want 42", val)
	}

	// Intermediate path should also work.
	aVal, ok := s.Get("a.b")
	if !ok {
		t.Fatal("expected intermediate path to exist")
	}
	m, ok := aVal.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", aVal)
	}
	if m["c"] != 42 {
		t.Errorf("got %v, want 42", m["c"])
	}
}

func TestToggle(t *testing.T) {
	s := New()
	s.Set("enabled", true)

	if err := s.Toggle("enabled"); err != nil {
		t.Fatal(err)
	}
	val, _ := s.Get("enabled")
	if val != false {
		t.Errorf("got %v, want false", val)
	}

	if err := s.Toggle("enabled"); err != nil {
		t.Fatal(err)
	}
	val, _ = s.Get("enabled")
	if val != true {
		t.Errorf("got %v, want true", val)
	}
}

func TestToggleMissingPath(t *testing.T) {
	s := New()
	err := s.Toggle("missing")
	if err == nil {
		t.Fatal("expected error for missing path")
	}
}

func TestToggleNonBool(t *testing.T) {
	s := New()
	s.Set("count", 42)
	err := s.Toggle("count")
	if err == nil {
		t.Fatal("expected error for non-bool")
	}
}

func TestSubscribe(t *testing.T) {
	s := New()

	var got []string
	unsub := s.Subscribe("user", func(key string, _ any) {
		got = append(got, key)
	})

	s.Set("user.name", "Alice")
	s.Set("user.email", "alice@example.com")
	s.Set("other.key", "ignored")

	if len(got) != 2 {
		t.Fatalf("expected 2 notifications, got %d: %v", len(got), got)
	}
	if got[0] != "user.name" || got[1] != "user.email" {
		t.Errorf("unexpected keys: %v", got)
	}

	unsub()
	s.Set("user.age", 30)
	if len(got) != 2 {
		t.Errorf("expected no more notifications after unsub, got %d", len(got))
	}
}

func TestSubscribeEmpty(t *testing.T) {
	s := New()
	var count int
	unsub := s.Subscribe("", func(_ string, _ any) {
		count++
	})
	defer unsub()

	s.Set("anything", "value")
	if count != 1 {
		t.Errorf("expected 1 notification with empty prefix, got %d", count)
	}
}

func TestSnapshot(t *testing.T) {
	s := New()
	s.Set("x", 1)
	s.Set("y.z", 2)

	snap := s.Snapshot()
	snap["x"] = 999

	val, _ := s.Get("x")
	if val != 1 {
		t.Errorf("snapshot mutation leaked: got %v, want 1", val)
	}
}

func TestLoad(t *testing.T) {
	s := New()
	s.Set("old", "value")

	s.Load(map[string]any{
		"new": "data",
	})

	_, ok := s.Get("old")
	if ok {
		t.Error("expected old key to be gone after Load")
	}
	val, ok := s.Get("new")
	if !ok {
		t.Fatal("expected new key to exist")
	}
	if val != "data" {
		t.Errorf("got %v, want data", val)
	}
}

func TestConcurrentAccess(t *testing.T) {
	s := New()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			s.Set("counter", i)
			s.Get("counter")
		}(i)
	}
	wg.Wait()

	_, ok := s.Get("counter")
	if !ok {
		t.Error("expected counter to exist")
	}
}

func TestToggleNotifies(t *testing.T) {
	s := New()
	s.Set("flag", true)

	var notified bool
	unsub := s.Subscribe("flag", func(_ string, _ any) {
		notified = true
	})
	defer unsub()

	if err := s.Toggle("flag"); err != nil {
		t.Fatal(err)
	}
	if !notified {
		t.Error("expected notification from Toggle")
	}
}

func TestAsContext(t *testing.T) {
	s := New()
	s.Set("key", "value")

	ctx := s.AsContext()
	ctx["key"] = "mutated"

	val, _ := s.Get("key")
	if val != "value" {
		t.Errorf("AsContext returned live reference, mutation leaked")
	}
}
