package core

import (
	"context"

	"github.com/kinjaze/git-worktree-manager/internal/jsonapi"
)

type MergeBackResult struct {
	TargetWorktree string `json:"targetWorktree"`
	TargetBranch   string `json:"targetBranch"`
	WorktreeBranch string `json:"worktreeBranch"`
}

func (m Manager) MergeBack(ctx context.Context, selector string) (MergeBackResult, error) {
	return m.MergeBackWithProgress(ctx, selector, noopProgress)
}

func (m Manager) MergeBackWithProgress(ctx context.Context, selector string, progress ProgressFunc) (MergeBackResult, error) {
	if progress == nil {
		progress = noopProgress
	}
	progress(1, 7, "Resolve worktree")
	record, err := m.resolveRecord(selector)
	if err != nil {
		return MergeBackResult{}, err
	}
	progress(2, 7, "Fetch source remote")
	if err := m.git.Fetch(ctx, record.SourceRepoPath, record.SourceRemote); err != nil {
		return MergeBackResult{}, err
	}
	progress(3, 7, "Ensure target branch")
	if !m.git.BranchExists(ctx, record.SourceRepoPath, record.TargetLocalBranch) {
		if err := m.git.CreateBranch(ctx, record.SourceRepoPath, record.TargetLocalBranch, record.SourceRemoteBranch); err != nil {
			return MergeBackResult{}, err
		}
	}
	progress(4, 7, "Locate target worktree")
	targetWorktree := record.SourceRepoPath
	worktrees, err := m.git.Worktrees(ctx, record.SourceRepoPath)
	if err == nil {
		for _, worktree := range worktrees {
			if worktree.Branch == record.TargetLocalBranch {
				targetWorktree = worktree.Path
				break
			}
		}
	}
	if targetWorktree == record.SourceRepoPath {
		if err := m.git.Checkout(ctx, targetWorktree, record.TargetLocalBranch); err != nil {
			return MergeBackResult{}, err
		}
	}
	progress(5, 7, "Check target status")
	dirty, err := m.git.IsDirty(ctx, targetWorktree)
	if err != nil {
		return MergeBackResult{}, err
	}
	if dirty {
		return MergeBackResult{}, NewError(jsonapi.ErrTargetDirty, "target worktree has uncommitted changes: %s", targetWorktree)
	}
	progress(6, 7, "Merge with --no-ff")
	if err := m.git.Merge(ctx, targetWorktree, record.WorktreeBranch, true); err != nil {
		if isMergeConflict(err) {
			files, _ := m.git.ConflictedFiles(ctx, targetWorktree)
			conflict := NewError(jsonapi.ErrMergeConflict, "merge conflict while merging back")
			conflict.Data = conflictData("merge-back", targetWorktree, record.Path, record.TargetLocalBranch, record.WorktreeBranch, files)
			return MergeBackResult{}, conflict
		}
		return MergeBackResult{}, err
	}
	record.UpdatedAt = now()
	progress(7, 7, "Update metadata")
	if err := m.store.Upsert(record); err != nil {
		return MergeBackResult{}, err
	}
	return MergeBackResult{TargetWorktree: targetWorktree, TargetBranch: record.TargetLocalBranch, WorktreeBranch: record.WorktreeBranch}, nil
}
