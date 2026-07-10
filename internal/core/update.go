package core

import (
	"context"

	"github.com/kinjaze/git-worktree-manager/internal/jsonapi"
)

type UpdateResult struct {
	Record any `json:"record,omitempty"`
}

func (m Manager) Update(ctx context.Context, selector string) (UpdateResult, error) {
	return m.UpdateWithProgress(ctx, selector, noopProgress)
}

func (m Manager) UpdateWithProgress(ctx context.Context, selector string, progress ProgressFunc) (UpdateResult, error) {
	if progress == nil {
		progress = noopProgress
	}
	progress(1, 4, "Resolve worktree")
	record, err := m.resolveRecord(selector)
	if err != nil {
		return UpdateResult{}, err
	}
	progress(2, 4, "Fetch source remote")
	if err := m.git.Fetch(ctx, record.Path, record.SourceRemote); err != nil {
		return UpdateResult{}, err
	}
	progress(3, 4, "Merge source branch")
	if err := m.git.Merge(ctx, record.Path, record.SourceRemoteBranch, false); err != nil {
		if isMergeConflict(err) {
			files, _ := m.git.ConflictedFiles(ctx, record.Path)
			conflict := NewError(jsonapi.ErrMergeConflict, "merge conflict while updating worktree")
			conflict.Data = conflictData("update", record.Path, record.Path, record.TargetLocalBranch, record.WorktreeBranch, files)
			return UpdateResult{}, conflict
		}
		return UpdateResult{}, err
	}
	record.UpdatedAt = now()
	record.LastKnownHead, _ = m.git.Head(ctx, record.Path)
	progress(4, 4, "Update metadata")
	if err := m.store.Upsert(record); err != nil {
		return UpdateResult{}, err
	}
	return UpdateResult{Record: record}, nil
}
