package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// RunEnv holds environment variables for --run command execution.
type RunEnv struct {
	SourceDir   string
	Branch      string
	Project     string
	WorktreeDir string
}

// Run executes a shell command with the given environment.
func Run(ctx context.Context, command string, env RunEnv) error {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = env.WorktreeDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = append(os.Environ(),
		"WT_SOURCE_DIR="+env.SourceDir,
		"WT_BRANCH="+env.Branch,
		"WT_PROJECT="+env.Project,
		"WT_WORKTREE_DIR="+env.WorktreeDir,
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("--run command failed: %w", err)
	}
	return nil
}
