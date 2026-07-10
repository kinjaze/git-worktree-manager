package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type Runner struct{}

type Result struct {
	Stdout string
	Stderr string
}

type CommandError struct {
	Args   []string
	Stdout string
	Stderr string
	Err    error
}

func (e CommandError) Error() string {
	message := strings.TrimSpace(e.Stderr)
	if message == "" {
		message = strings.TrimSpace(e.Stdout)
	}
	if message == "" && e.Err != nil {
		message = e.Err.Error()
	}
	return fmt.Sprintf("git %s failed: %s", strings.Join(e.Args, " "), message)
}

func NewRunner() Runner {
	return Runner{}
}

func (r Runner) Run(ctx context.Context, dir string, args ...string) (Result, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return Result{Stdout: stdout.String(), Stderr: stderr.String()}, CommandError{Args: args, Stdout: stdout.String(), Stderr: stderr.String(), Err: err}
	}
	return Result{Stdout: stdout.String(), Stderr: stderr.String()}, nil
}
