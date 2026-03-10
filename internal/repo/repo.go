package repo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/glennr/wt/internal/worktree"
)

// Resolve determines the source repo path and project name.
// Resolution order: -p flag -> cwd worktree -> cwd git repo -> error.
func Resolve(ctx context.Context, projectFlag string) (string, string, error) {
	if projectFlag != "" {
		return resolveProjectFlag(ctx, projectFlag)
	}

	// Try to detect if we're inside a worktree and find the main repo
	mainWorktree, err := gitMainWorktree(ctx)
	if err == nil && mainWorktree != "" {
		return mainWorktree, filepath.Base(mainWorktree), nil
	}

	// Try cwd git repo
	topLevel, err := gitTopLevel(ctx)
	if err != nil {
		return "", "", errors.New("not in a git repo; use -p to specify a project")
	}
	return topLevel, filepath.Base(topLevel), nil
}

// resolveProjectFlag handles -p/--project flag resolution.
// Contains /, ., or ~ → path; bare name → ~/worktrees/<name>/.
func resolveProjectFlag(ctx context.Context, flag string) (string, string, error) {
	if isPath(flag) {
		expanded := expandTilde(flag)
		abs, err := filepath.Abs(expanded)
		if err != nil {
			return "", "", fmt.Errorf("resolving -p path: %w", err)
		}
		return abs, filepath.Base(abs), nil
	}

	// Bare name: look up ~/worktrees/<name>/
	home, _ := os.UserHomeDir()
	projectDir := filepath.Join(home, "worktrees", flag)
	if _, err := os.Stat(projectDir); err != nil {
		return "", "", fmt.Errorf("project %q not found at %s", flag, projectDir)
	}

	// Find source repo from an existing worktree
	repoPath, err := findSourceRepo(ctx, projectDir)
	if err != nil {
		return "", "", fmt.Errorf("could not determine source repo for project %q: %w", flag, err)
	}
	return repoPath, flag, nil
}

// isPath returns true if the string looks like a file path rather than a bare project name.
func isPath(s string) bool {
	return strings.ContainsAny(s, "/.~")
}

// expandTilde replaces a leading ~ with the user's home directory.
func expandTilde(path string) string {
	if strings.HasPrefix(path, "~/") || path == "~" {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[1:])
	}
	return path
}

// findSourceRepo finds the source repo by looking at an existing worktree's git metadata.
func findSourceRepo(ctx context.Context, projectDir string) (string, error) {
	entries := worktree.ListWorktrees(projectDir)
	if len(entries) == 0 {
		return "", errors.New("no existing worktrees found")
	}
	return gitMainWorktreeFrom(ctx, entries[0].Path)
}

func gitTopLevel(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func gitMainWorktree(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "git", "worktree", "list", "--porcelain").Output()
	if err != nil {
		return "", err
	}
	return parseMainWorktree(string(out))
}

func gitMainWorktreeFrom(ctx context.Context, dir string) (string, error) {
	out, err := exec.CommandContext(ctx, "git", "-C", dir, "worktree", "list", "--porcelain").Output()
	if err != nil {
		return "", err
	}
	return parseMainWorktree(string(out))
}

func parseMainWorktree(porcelain string) (string, error) {
	for line := range strings.SplitSeq(porcelain, "\n") {
		if after, ok := strings.CutPrefix(line, "worktree "); ok {
			return after, nil
		}
	}
	return "", errors.New("no worktree found")
}
