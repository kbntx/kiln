package runner

import (
	"sync"
	"testing"
	"time"

	"github.com/kbntx/kiln/internal/discovery"
)

func TestRunStore_Create(t *testing.T) {
	store := NewRunStore()
	run := &Run{
		Owner:    "acme",
		Repo:     "infra",
		PRNumber: 1,
		Status:   RunStatusPending,
	}

	store.Create(run)

	if run.ID == "" {
		t.Fatal("expected run to have an ID after Create")
	}
	if run.CreatedAt.IsZero() {
		t.Fatal("expected run to have a CreatedAt after Create")
	}
	if time.Since(run.CreatedAt) > time.Second {
		t.Fatalf("CreatedAt should be recent, got %v", run.CreatedAt)
	}
}

func TestRunStore_Get(t *testing.T) {
	store := NewRunStore()
	run := &Run{Owner: "acme", Repo: "infra", Status: RunStatusPending}
	store.Create(run)

	got := store.Get(run.ID)
	if got == nil {
		t.Fatal("expected to get the run back")
	}
	if got.ID != run.ID {
		t.Fatalf("expected ID %q, got %q", run.ID, got.ID)
	}
	if got.Owner != "acme" {
		t.Fatalf("expected Owner %q, got %q", "acme", got.Owner)
	}
}

func TestRunStore_GetNonExistent(t *testing.T) {
	store := NewRunStore()

	got := store.Get("does-not-exist")
	if got != nil {
		t.Fatal("expected nil for non-existent run")
	}
}

func TestRunStore_Update(t *testing.T) {
	store := NewRunStore()
	run := &Run{Owner: "acme", Repo: "infra", Status: RunStatusPending}
	store.Create(run)

	store.Update(run.ID, func(r *Run) {
		r.Status = RunStatusRunning
	})

	got := store.Get(run.ID)
	if got.Status != RunStatusRunning {
		t.Fatalf("expected status %q, got %q", RunStatusRunning, got.Status)
	}
}

func TestRunStore_UpdateNonExistent(t *testing.T) {
	store := NewRunStore()
	// Should not panic when updating a non-existent run.
	store.Update("nope", func(r *Run) {
		r.Status = RunStatusFailed
	})
}

func TestRunStore_FindConfig(t *testing.T) {
	store := NewRunStore()
	run := &Run{Owner: "acme", Repo: "infra", PRNumber: 42, HeadSHA: "abc123", Status: RunStatusSuccess}
	store.Create(run)

	cfg := &discovery.Config{
		Projects: []discovery.ProjectConfig{
			{Name: "test", Dir: ".", Engine: "terraform", Stacks: []string{"default"}},
		},
	}
	store.Update(run.ID, func(r *Run) {
		r.Config = cfg
	})

	got := store.FindConfig("acme", "infra", "abc123")
	if got == nil {
		t.Fatal("expected to find config")
	}
	if len(got.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(got.Projects))
	}

	got = store.FindConfig("acme", "infra", "def456")
	if got != nil {
		t.Fatalf("expected nil config for different SHA, got %+v", got)
	}
}

func TestRunStore_ConcurrentCreateGet(t *testing.T) {
	store := NewRunStore()
	const n = 100
	ids := make([]string, n)
	var mu sync.Mutex

	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			run := &Run{Owner: "acme", Repo: "infra", PRNumber: i, Status: RunStatusPending}
			store.Create(run)
			mu.Lock()
			ids[i] = run.ID
			mu.Unlock()
		}(i)
	}
	wg.Wait()

	// Verify all runs can be retrieved.
	for i, id := range ids {
		got := store.Get(id)
		if got == nil {
			t.Fatalf("run %d (id %q) not found", i, id)
		}
		if got.PRNumber != i {
			t.Fatalf("expected PRNumber %d, got %d", i, got.PRNumber)
		}
	}
}
