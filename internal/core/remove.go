package core

import (
	"context"
	"os"

	"github.com/qinbin/git-worktree-manager/internal/jsonapi"
)

type RemoveOptions struct {
	Selector     string
	Force        bool
	MetadataOnly bool
}

type RemoveResult struct {
	ID            string `json:"id"`
	Path          string `json:"path"`
	Branch        string `json:"branch"`
	BranchDeleted bool   `json:"branchDeleted"`
}

func (m Manager) Remove(ctx context.Context, options RemoveOptions) (RemoveResult, error) {
	return m.RemoveWithProgress(ctx, options, noopProgress)
}

func (m Manager) RemoveWithProgress(ctx context.Context, options RemoveOptions, progress ProgressFunc) (RemoveResult, error) {
	if progress == nil {
		progress = noopProgress
	}
	progress(1, 5, "Resolve worktree")
	record, err := m.resolveRecord(options.Selector)
	if err != nil {
		return RemoveResult{}, err
	}
	if options.MetadataOnly {
		progress(5, 5, "Remove metadata")
		if err := m.store.Remove(record.ID); err != nil {
			return RemoveResult{}, err
		}
		return RemoveResult{ID: record.ID, Path: record.Path, Branch: record.WorktreeBranch}, nil
	}
	progress(2, 5, "Check worktree status")
	if _, err := os.Stat(record.Path); err == nil {
		dirty, err := m.git.IsDirty(ctx, record.Path)
		if err != nil {
			return RemoveResult{}, err
		}
		if dirty && !options.Force {
			return RemoveResult{}, NewError(jsonapi.ErrWorktreeDirty, "worktree has uncommitted changes: %s", record.Path)
		}
		progress(3, 5, "Remove git worktree")
		if err := m.git.WorktreeRemove(ctx, record.SourceRepoPath, record.Path, options.Force); err != nil {
			return RemoveResult{}, err
		}
	}
	progress(4, 5, "Delete worktree branch")
	if err := m.git.DeleteBranch(ctx, record.SourceRepoPath, record.WorktreeBranch, options.Force); err != nil {
		return RemoveResult{}, err
	}
	progress(5, 5, "Remove metadata")
	if err := m.store.Remove(record.ID); err != nil {
		return RemoveResult{}, err
	}
	return RemoveResult{ID: record.ID, Path: record.Path, Branch: record.WorktreeBranch, BranchDeleted: true}, nil
}
