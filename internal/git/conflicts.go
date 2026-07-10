package git

import (
	"context"
	"strings"
)

func (r Runner) ConflictedFiles(ctx context.Context, dir string) ([]string, error) {
	result, err := r.Run(ctx, dir, "diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return nil, err
	}
	var files []string
	for _, line := range strings.Split(result.Stdout, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}
