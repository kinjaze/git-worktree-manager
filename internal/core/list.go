package core

import (
	"context"
	"os"

	"github.com/kinjaze/git-worktree-manager/internal/metadata"
)

type ListResult struct {
	Worktrees []metadata.Record `json:"worktrees"`
}

func (m Manager) List(ctx context.Context) (ListResult, error) {
	file, err := m.store.Load()
	if err != nil {
		return ListResult{}, err
	}
	records := make([]metadata.Record, 0, len(file.Worktrees))
	for _, record := range file.Worktrees {
		status := "active"
		if _, err := os.Stat(record.Path); err != nil {
			status = "missing_path"
		} else if files, err := m.git.ConflictedFiles(ctx, record.Path); err == nil && len(files) > 0 {
			status = "conflict"
		} else if dirty, err := m.git.IsDirty(ctx, record.Path); err == nil && dirty {
			status = "dirty"
		} else if err != nil {
			status = "unknown"
		}
		record.Status = status
		records = append(records, record)
	}
	return ListResult{Worktrees: records}, nil
}
