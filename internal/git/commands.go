package git

import (
	"context"
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
