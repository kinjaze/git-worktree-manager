package core

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	gitpkg "github.com/kinjaze/git-worktree-manager/internal/git"
	"github.com/kinjaze/git-worktree-manager/internal/metadata"
)

func TestRemovePrunesDeletedWorktree(t *testing.T) {
	ctx := context.Background()
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	worktreePath := filepath.Join(tmp, "worktree")
	metadataPath := filepath.Join(tmp, "metadata.json")

	runGit(t, "", "init", repo)
	runGit(t, repo, "config", "user.name", "Test User")
	runGit(t, repo, "config", "user.email", "test@example.com")
	if err := os.WriteFile(filepath.Join(repo, "file.txt"), []byte("initial\n"), 0o644); err != nil {
		t.Fatalf("write seed file: %v", err)
	}
	runGit(t, repo, "add", "file.txt")
	runGit(t, repo, "commit", "-m", "initial")
	runGit(t, repo, "worktree", "add", "-b", "feature/test", worktreePath)
	if err := os.RemoveAll(worktreePath); err != nil {
		t.Fatalf("remove worktree path: %v", err)
	}

	store := metadata.NewJSONStore(metadataPath)
	if err := store.Upsert(metadata.Record{
		ID:             "test",
		Name:           "test",
		SourceRepoPath: repo,
		WorktreeBranch: "feature/test",
		Path:           worktreePath,
	}); err != nil {
		t.Fatalf("seed metadata: %v", err)
	}

	manager := NewManager(gitpkg.NewRunner(), store)
	if _, err := manager.Remove(ctx, RemoveOptions{Selector: "test"}); err != nil {
		t.Fatalf("Remove returned error: %v", err)
	}

	file, err := store.Load()
	if err != nil {
		t.Fatalf("load metadata: %v", err)
	}
	if len(file.Worktrees) != 0 {
		t.Fatalf("expected metadata removed, got %+v", file.Worktrees)
	}
	if gitpkg.NewRunner().BranchExists(ctx, repo, "feature/test") {
		t.Fatal("expected worktree branch to be deleted")
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, output)
	}
}
