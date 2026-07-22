package git

import (
	"context"
	"path/filepath"
	"strings"
)

func (r Runner) IsGitRepository(ctx context.Context, dir string) bool {
	_, err := r.Run(ctx, dir, "rev-parse", "--git-dir")
	return err == nil
}

func (r Runner) Fetch(ctx context.Context, dir string, remote string) error {
	_, err := r.Run(ctx, dir, "fetch", remote)
	return err
}

func (r Runner) VerifyRef(ctx context.Context, dir string, ref string) error {
	_, err := r.Run(ctx, dir, "rev-parse", "--verify", ref)
	return err
}

func (r Runner) WorktreeAdd(ctx context.Context, repo string, branch string, path string, ref string) error {
	_, err := r.Run(ctx, repo, "worktree", "add", "-b", branch, path, ref)
	return err
}

func (r Runner) WorktreeRemove(ctx context.Context, repo string, path string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)
	_, err := r.Run(ctx, repo, args...)
	return err
}

func (r Runner) WorktreePrune(ctx context.Context, repo string) error {
	_, err := r.Run(ctx, repo, "worktree", "prune", "--expire", "now")
	return err
}

func (r Runner) WorktreeRepair(ctx context.Context, repo string, path string) error {
	_, err := r.Run(ctx, repo, "worktree", "repair", path)
	return err
}

func (r Runner) Merge(ctx context.Context, dir string, ref string, noFF bool) error {
	args := []string{"merge"}
	if noFF {
		args = append(args, "--no-ff")
	}
	args = append(args, ref)
	_, err := r.Run(ctx, dir, args...)
	return err
}

func (r Runner) BranchExists(ctx context.Context, repo string, branch string) bool {
	_, err := r.Run(ctx, repo, "rev-parse", "--verify", "refs/heads/"+branch)
	return err == nil
}

func (r Runner) CreateBranch(ctx context.Context, repo string, branch string, ref string) error {
	_, err := r.Run(ctx, repo, "branch", branch, ref)
	return err
}

func (r Runner) DeleteBranch(ctx context.Context, repo string, branch string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}
	_, err := r.Run(ctx, repo, "branch", flag, branch)
	return err
}

func (r Runner) Checkout(ctx context.Context, repo string, branch string) error {
	_, err := r.Run(ctx, repo, "checkout", branch)
	return err
}

func (r Runner) Worktrees(ctx context.Context, repo string) ([]Worktree, error) {
	result, err := r.Run(ctx, repo, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	return ParseWorktreeList(result.Stdout), nil
}

func (r Runner) FindWorktree(ctx context.Context, repo string, path string) (Worktree, bool, error) {
	worktrees, err := r.Worktrees(ctx, repo)
	if err != nil {
		return Worktree{}, false, err
	}
	path, err = canonicalPath(path)
	if err != nil {
		return Worktree{}, false, err
	}
	for _, worktree := range worktrees {
		worktreePath, err := canonicalPath(worktree.Path)
		if err != nil {
			return Worktree{}, false, err
		}
		if worktreePath == path {
			return worktree, true, nil
		}
	}
	return Worktree{}, false, nil
}

func canonicalPath(path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		return resolved, nil
	}
	return path, nil
}

func (r Runner) IsDirty(ctx context.Context, dir string) (bool, error) {
	result, err := r.Run(ctx, dir, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(result.Stdout) != "", nil
}

func (r Runner) Head(ctx context.Context, dir string) (string, error) {
	result, err := r.Run(ctx, dir, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(result.Stdout), nil
}

func (r Runner) CurrentBranch(ctx context.Context, dir string) (string, error) {
	result, err := r.Run(ctx, dir, "branch", "--show-current")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(result.Stdout), nil
}

func (r Runner) CommonDir(ctx context.Context, dir string) (string, error) {
	result, err := r.Run(ctx, dir, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(result.Stdout), nil
}
