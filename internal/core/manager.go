package core

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"path/filepath"
	"strings"
	"time"

	gitpkg "github.com/kinjaze/git-worktree-manager/internal/git"
	"github.com/kinjaze/git-worktree-manager/internal/jsonapi"
	"github.com/kinjaze/git-worktree-manager/internal/metadata"
)

type Manager struct {
	git   gitpkg.Runner
	store metadata.Store
}

func NewManager(git gitpkg.Runner, store metadata.Store) Manager {
	return Manager{git: git, store: store}
}

func stableID(parts ...string) string {
	hash := sha1.Sum([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(hash[:])[:12]
}

func now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func absPath(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	return filepath.Abs(path)
}

func (m Manager) resolveRecord(selector string) (metadata.Record, error) {
	file, err := m.store.Load()
	if err != nil {
		return metadata.Record{}, err
	}
	absolute, _ := filepath.Abs(selector)
	for _, record := range file.Worktrees {
		if record.ID == selector || record.Name == selector || record.Path == selector || record.Path == absolute {
			return record, nil
		}
	}
	return metadata.Record{}, NewError(jsonapi.ErrWorktreeNotFound, "worktree not found: %s", selector)
}

func isMergeConflict(err error) bool {
	if commandErr, ok := err.(gitpkg.CommandError); ok {
		text := strings.ToLower(commandErr.Stdout + "\n" + commandErr.Stderr)
		return strings.Contains(text, "conflict") || strings.Contains(text, "automatic merge failed") || strings.Contains(text, "fix conflicts")
	}
	return false
}

func conflictData(operation string, targetWorktree string, sourceWorktree string, targetBranch string, worktreeBranch string, files []string) map[string]any {
	return map[string]any{
		"operation":       operation,
		"targetWorktree":  targetWorktree,
		"sourceWorktree":  sourceWorktree,
		"targetBranch":    targetBranch,
		"worktreeBranch":  worktreeBranch,
		"conflictedFiles": files,
		"nextSteps": []string{
			"Resolve conflicted files",
			"git add <files>",
			"git merge --continue",
			"git merge --abort",
		},
	}
}

func (m Manager) Git() gitpkg.Runner {
	return m.git
}

func (m Manager) Store() metadata.Store {
	return m.store
}

func Context() context.Context {
	return context.Background()
}
