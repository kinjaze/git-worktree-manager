package core

import (
	"context"

	gitpkg "github.com/qinbin/git-worktree-manager/internal/git"
	"github.com/qinbin/git-worktree-manager/internal/jsonapi"
	"github.com/qinbin/git-worktree-manager/internal/metadata"
)

type CreateOptions struct {
	Name   string
	Repo   string
	Source string
	Branch string
	Path   string
}

type CreateResult struct {
	Record metadata.Record `json:"record"`
}

func (m Manager) Create(ctx context.Context, options CreateOptions) (CreateResult, error) {
	return m.CreateWithProgress(ctx, options, noopProgress)
}

func (m Manager) CreateWithProgress(ctx context.Context, options CreateOptions, progress ProgressFunc) (CreateResult, error) {
	if progress == nil {
		progress = noopProgress
	}
	progress(1, 5, "Validate source repo")
	if options.Name == "" || options.Repo == "" || options.Source == "" || options.Branch == "" || options.Path == "" {
		return CreateResult{}, NewError(jsonapi.ErrInvalidArgument, "name, --repo, --source, --branch, and --path are required")
	}
	repo, err := absPath(options.Repo)
	if err != nil {
		return CreateResult{}, err
	}
	path, err := absPath(options.Path)
	if err != nil {
		return CreateResult{}, err
	}
	if !m.git.IsGitRepository(ctx, repo) {
		return CreateResult{}, NewError(jsonapi.ErrNotGitRepository, "not a git repository: %s", repo)
	}
	remoteBranch, err := gitpkg.ParseRemoteBranch(options.Source)
	if err != nil {
		return CreateResult{}, NewError(jsonapi.ErrInvalidArgument, "%s", err.Error())
	}
	progress(2, 5, "Fetch source remote")
	if err := m.git.Fetch(ctx, repo, remoteBranch.Remote); err != nil {
		return CreateResult{}, err
	}
	progress(3, 5, "Verify source branch")
	if err := m.git.VerifyRef(ctx, repo, remoteBranch.Ref); err != nil {
		return CreateResult{}, NewError(jsonapi.ErrSourceRefNotFound, "source ref not found: %s", remoteBranch.Ref)
	}
	progress(4, 5, "Create worktree")
	if err := m.git.WorktreeAdd(ctx, repo, options.Branch, path, remoteBranch.Ref); err != nil {
		return CreateResult{}, err
	}
	head, _ := m.git.Head(ctx, path)
	timestamp := now()
	record := metadata.Record{
		ID:                 stableID(repo, path, options.Branch),
		Name:               options.Name,
		SourceRepoPath:     repo,
		SourceRemoteBranch: remoteBranch.Ref,
		SourceRemote:       remoteBranch.Remote,
		SourceBranchName:   remoteBranch.Branch,
		TargetLocalBranch:  remoteBranch.Branch,
		WorktreeBranch:     options.Branch,
		Path:               path,
		CreatedAt:          timestamp,
		UpdatedAt:          timestamp,
		LastKnownHead:      head,
		Status:             "active",
	}
	progress(5, 5, "Write metadata")
	if err := m.store.Upsert(record); err != nil {
		return CreateResult{}, err
	}
	return CreateResult{Record: record}, nil
}
