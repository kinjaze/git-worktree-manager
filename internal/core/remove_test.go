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

func TestRemoveMovedWorktreeByNewPath(t *testing.T) {
	ctx := context.Background()
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	oldPath := filepath.Join(tmp, "fixBug")
	newPath := filepath.Join(tmp, "web", "fixBug")
	metadataPath := filepath.Join(tmp, "metadata.json")

	initRepo(t, repo)
	runGit(t, repo, "worktree", "add", "-b", "feature/test", oldPath)
	if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
		t.Fatalf("create new parent: %v", err)
	}
	if err := os.Rename(oldPath, newPath); err != nil {
		t.Fatalf("move worktree: %v", err)
	}

	store := metadata.NewJSONStore(metadataPath)
	if err := store.Upsert(metadata.Record{
		ID:             "test",
		Name:           "test",
		SourceRepoPath: repo,
		WorktreeBranch: "feature/test",
		Path:           oldPath,
	}); err != nil {
		t.Fatalf("seed metadata: %v", err)
	}

	manager := NewManager(gitpkg.NewRunner(), store)
	if _, err := manager.Remove(ctx, RemoveOptions{Selector: newPath}); err != nil {
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

func TestUpdateMovedWorktreeByNewPath(t *testing.T) {
	ctx := context.Background()
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	oldPath := filepath.Join(tmp, "fixBug")
	newPath := filepath.Join(tmp, "web", "fixBug")
	metadataPath := filepath.Join(tmp, "metadata.json")

	initRepo(t, repo)
	runGit(t, repo, "worktree", "add", "-b", "feature/test", oldPath)
	if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
		t.Fatalf("create new parent: %v", err)
	}
	if err := os.Rename(oldPath, newPath); err != nil {
		t.Fatalf("move worktree: %v", err)
	}

	store := metadata.NewJSONStore(metadataPath)
	if err := store.Upsert(metadata.Record{
		ID:                 "test",
		Name:               "test",
		SourceRepoPath:     repo,
		SourceRemote:       ".",
		SourceRemoteBranch: "main",
		WorktreeBranch:     "feature/test",
		Path:               oldPath,
	}); err != nil {
		t.Fatalf("seed metadata: %v", err)
	}

	manager := NewManager(gitpkg.NewRunner(), store)
	if _, err := manager.Update(ctx, newPath); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	file, err := store.Load()
	if err != nil {
		t.Fatalf("load metadata: %v", err)
	}
	if len(file.Worktrees) != 1 {
		t.Fatalf("expected one record, got %+v", file.Worktrees)
	}
	if file.Worktrees[0].Path != newPath {
		t.Fatalf("expected path %q, got %q", newPath, file.Worktrees[0].Path)
	}
	if _, found, err := gitpkg.NewRunner().FindWorktree(ctx, repo, newPath); err != nil || !found {
		t.Fatalf("expected git worktree at new path, found=%v err=%v", found, err)
	}
}

func TestUpdateMovedRepoAndWorktreeByName(t *testing.T) {
	ctx := context.Background()
	tmp := t.TempDir()
	oldRepo := filepath.Join(tmp, "obcp", "obcp-web")
	oldPath := filepath.Join(tmp, "obcp", "fixBug")
	newRepo := filepath.Join(tmp, "obcp", "web", "obcp-web")
	newPath := filepath.Join(tmp, "obcp", "web", "fixBug")
	metadataPath := filepath.Join(tmp, "metadata.json")

	initRepo(t, oldRepo)
	runGit(t, oldRepo, "worktree", "add", "-b", "fixBug", oldPath)
	if err := os.MkdirAll(filepath.Dir(newRepo), 0o755); err != nil {
		t.Fatalf("create new parent: %v", err)
	}
	if err := os.Rename(oldRepo, newRepo); err != nil {
		t.Fatalf("move repo: %v", err)
	}
	if err := os.Rename(oldPath, newPath); err != nil {
		t.Fatalf("move worktree: %v", err)
	}

	store := metadata.NewJSONStore(metadataPath)
	if err := store.Upsert(metadata.Record{
		ID:                 "test",
		Name:               "fixBug",
		SourceRepoPath:     oldRepo,
		SourceRemote:       ".",
		SourceRemoteBranch: "main",
		WorktreeBranch:     "fixBug",
		Path:               oldPath,
	}); err != nil {
		t.Fatalf("seed metadata: %v", err)
	}

	manager := NewManager(gitpkg.NewRunner(), store)
	if _, err := manager.Update(ctx, "fixBug"); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	file, err := store.Load()
	if err != nil {
		t.Fatalf("load metadata: %v", err)
	}
	if len(file.Worktrees) != 1 {
		t.Fatalf("expected one record, got %+v", file.Worktrees)
	}
	if !samePath(file.Worktrees[0].SourceRepoPath, newRepo) {
		t.Fatalf("expected repo %q, got %q", newRepo, file.Worktrees[0].SourceRepoPath)
	}
	if !samePath(file.Worktrees[0].Path, newPath) {
		t.Fatalf("expected path %q, got %q", newPath, file.Worktrees[0].Path)
	}
}

func TestRemovePrunesDeletedWorktree(t *testing.T) {
	ctx := context.Background()
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	worktreePath := filepath.Join(tmp, "worktree")
	metadataPath := filepath.Join(tmp, "metadata.json")

	initRepo(t, repo)
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

func initRepo(t *testing.T, repo string) {
	t.Helper()
	runGit(t, "", "init", repo)
	runGit(t, repo, "config", "user.name", "Test User")
	runGit(t, repo, "config", "user.email", "test@example.com")
	if err := os.WriteFile(filepath.Join(repo, "file.txt"), []byte("initial\n"), 0o644); err != nil {
		t.Fatalf("write seed file: %v", err)
	}
	runGit(t, repo, "add", "file.txt")
	runGit(t, repo, "commit", "-m", "initial")
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
