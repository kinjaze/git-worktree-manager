package metadata

import (
	"path/filepath"
	"testing"
)

func TestJSONStoreUpsertAndRemove(t *testing.T) {
	path := filepath.Join(t.TempDir(), "metadata.json")
	store := NewJSONStore(path)
	record := Record{ID: "abc", Name: "demo", Path: "/tmp/demo"}
	if err := store.Upsert(record); err != nil {
		t.Fatalf("Upsert returned error: %v", err)
	}
	file, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(file.Worktrees) != 1 || file.Worktrees[0].ID != "abc" {
		t.Fatalf("unexpected metadata: %+v", file)
	}
	if err := store.Remove("abc"); err != nil {
		t.Fatalf("Remove returned error: %v", err)
	}
	file, err = store.Load()
	if err != nil {
		t.Fatalf("Load after remove returned error: %v", err)
	}
	if len(file.Worktrees) != 0 {
		t.Fatalf("expected no worktrees, got %+v", file.Worktrees)
	}
}
