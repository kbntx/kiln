package runner

import (
	"sync"
	"testing"
	"time"
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

func TestRunStore_FindWorkDir(t *testing.T) {
	store := NewRunStore()
	run := &Run{Owner: "acme", Repo: "infra", PRNumber: 42, Status: RunStatusSuccess}
	store.Create(run)
	store.Update(run.ID, func(r *Run) {
		r.WorkDir = "/tmp/workspace/repo"
	})

	got := store.FindWorkDir("acme", "infra", 42)
	if got != "/tmp/workspace/repo" {
		t.Fatalf("expected workdir %q, got %q", "/tmp/workspace/repo", got)
	}

	got = store.FindWorkDir("acme", "infra", 99)
	if got != "" {
		t.Fatalf("expected empty workdir for unknown PR, got %q", got)
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
