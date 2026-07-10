package git

import "testing"

func TestParseWorktreeList(t *testing.T) {
	output := "worktree /repo\nHEAD abc\nbranch refs/heads/main\n\nworktree /repo/.worktrees/login\nHEAD def\nbranch refs/heads/feat/login\n"
	worktrees := ParseWorktreeList(output)
	if len(worktrees) != 2 {
		t.Fatalf("expected 2 worktrees, got %d", len(worktrees))
	}
	if worktrees[1].Path != "/repo/.worktrees/login" || worktrees[1].Branch != "feat/login" {
		t.Fatalf("unexpected second worktree: %+v", worktrees[1])
	}
}
