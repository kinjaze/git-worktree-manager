package core

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"os"
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

func (m Manager) resolveRecord(ctx context.Context, selector string) (metadata.Record, error) {
	file, err := m.store.Load()
	if err != nil {
		return metadata.Record{}, err
	}
	absolute, _ := filepath.Abs(selector)
	for _, record := range file.Worktrees {
		if record.ID == selector || record.Name == selector || record.Path == selector || samePath(record.Path, absolute) {
			return m.refreshRecordPath(ctx, record)
		}
	}
	if record, found, err := m.resolveRecordByCurrentWorktree(ctx, file.Worktrees, selector, absolute); err != nil || found {
		return record, err
	}
	return metadata.Record{}, NewError(jsonapi.ErrWorktreeNotFound, "worktree not found: %s", selector)
}

func (m Manager) refreshRecordPath(ctx context.Context, record metadata.Record) (metadata.Record, error) {
	worktrees, err := m.git.Worktrees(ctx, record.SourceRepoPath)
	if err != nil {
		return m.repairMovedRecord(ctx, record)
	}
	for _, worktree := range worktrees {
		if worktree.Branch != record.WorktreeBranch || worktree.Prunable {
			continue
		}
		if samePath(record.Path, worktree.Path) {
			return record, nil
		}
		record.Path = worktree.Path
		if err := m.store.Upsert(record); err != nil {
			return metadata.Record{}, err
		}
		return record, nil
	}
	return m.repairMovedRecord(ctx, record)
}

func (m Manager) repairMovedRecord(ctx context.Context, record metadata.Record) (metadata.Record, error) {
	if _, err := os.Stat(record.Path); err == nil {
		return record, nil
	}
	for _, repoPath := range nearbyRepoCandidates(record.SourceRepoPath) {
		for _, worktreePath := range nearbyWorktreeCandidates(record.Path) {
			if err := m.git.WorktreeRepair(ctx, repoPath, worktreePath); err != nil {
				continue
			}
			worktrees, err := m.git.Worktrees(ctx, repoPath)
			if err != nil {
				continue
			}
			for _, worktree := range worktrees {
				if worktree.Branch != record.WorktreeBranch || worktree.Prunable {
					continue
				}
				record.SourceRepoPath = repoPath
				record.Path = worktree.Path
				if err := m.store.Upsert(record); err != nil {
					return metadata.Record{}, err
				}
				return record, nil
			}
		}
	}
	return record, nil
}

func nearbyRepoCandidates(path string) []string {
	return nearbyMovedPathCandidates(path, func(candidate string) bool {
		_, err := os.Stat(filepath.Join(candidate, ".git"))
		return err == nil
	})
}

func nearbyWorktreeCandidates(path string) []string {
	return nearbyMovedPathCandidates(path, func(candidate string) bool {
		_, err := os.Stat(candidate)
		return err == nil
	})
}

func nearbyMovedPathCandidates(path string, exists func(string) bool) []string {
	parent := filepath.Dir(path)
	base := filepath.Base(path)
	var candidates []string
	seen := map[string]bool{}
	add := func(candidate string) {
		absolute, err := filepath.Abs(candidate)
		if err != nil || seen[absolute] || !exists(absolute) {
			return
		}
		seen[absolute] = true
		candidates = append(candidates, absolute)
	}
	add(path)
	if entries, err := os.ReadDir(parent); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				add(filepath.Join(parent, entry.Name(), base))
			}
		}
	}
	return candidates
}

func (m Manager) resolveRecordByCurrentWorktree(ctx context.Context, records []metadata.Record, selector string, path string) (metadata.Record, bool, error) {
	for _, record := range records {
		selectorBase := filepath.Base(path)
		if record.ID != selector && record.Name != selector && record.WorktreeBranch != selector && record.Name != selectorBase && record.WorktreeBranch != selectorBase {
			continue
		}
		repaired, err := m.repairMovedRecord(ctx, record)
		return repaired, true, err
	}
	branch, err := m.git.CurrentBranch(ctx, path)
	if err != nil || branch == "" {
		return metadata.Record{}, false, nil
	}
	commonDir, err := m.git.CommonDir(ctx, path)
	if err != nil {
		return metadata.Record{}, false, nil
	}
	for _, record := range records {
		if record.WorktreeBranch != branch {
			continue
		}
		for _, repoPath := range nearbyRepoCandidates(record.SourceRepoPath) {
			recordCommonDir, err := m.git.CommonDir(ctx, repoPath)
			if err != nil || !samePath(commonDir, recordCommonDir) {
				continue
			}
			if err := m.git.WorktreeRepair(ctx, repoPath, path); err != nil {
				return metadata.Record{}, true, err
			}
			record.SourceRepoPath = repoPath
			record.Path = path
			if err := m.store.Upsert(record); err != nil {
				return metadata.Record{}, true, err
			}
			return record, true, nil
		}
	}
	return metadata.Record{}, false, nil
}

func samePath(a string, b string) bool {
	return comparablePath(a) == comparablePath(b)
}

func comparablePath(path string) string {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err == nil {
		return resolved
	}
	return absolute
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
