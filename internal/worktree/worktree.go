package worktree

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Dir returns ~/worktrees/<project>/<branch>, preserving / in branch names as nested directories.
func Dir(project, branch string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "worktrees", project, branch)
}

// FindWorktree resolves a branch name or path to an existing worktree directory.
// Returns an error if the directory exists but has no .git marker (broken metadata).
func FindWorktree(project, branchOrPath string) (string, error) {
	if filepath.IsAbs(branchOrPath) {
		if _, err := os.Stat(filepath.Join(branchOrPath, ".git")); err == nil {
			return branchOrPath, nil
		}
		if _, err := os.Stat(branchOrPath); err == nil {
			return "", fmt.Errorf("directory %q exists but is not a valid worktree", branchOrPath)
		}
		return "", fmt.Errorf("worktree not found at %q", branchOrPath)
	}

	dir := Dir(project, branchOrPath)
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		return dir, nil
	}
	if _, err := os.Stat(dir); err == nil {
		return "", fmt.Errorf(
			"directory exists but is not a valid worktree for %q in project %q",
			branchOrPath,
			project,
		)
	}
	return "", fmt.Errorf("worktree not found for %q in project %q", branchOrPath, project)
}

// Entry represents a discovered worktree.
type Entry struct {
	Branch string
	Path   string
}

// ListWorktrees returns all worktree entries under a project directory.
// It walks the tree and finds directories with a .git marker.
func ListWorktrees(projectDir string) []Entry {
	var entries []Entry
	_ = filepath.WalkDir(projectDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || !d.IsDir() || path == projectDir {
			return walkErr
		}
		if _, serr := os.Stat(filepath.Join(path, ".git")); serr == nil {
			rel, _ := filepath.Rel(projectDir, path)
			entries = append(entries, Entry{Branch: rel, Path: path})
			return fs.SkipDir
		}
		return nil
	})
	return entries
}

// DetectBranch determines the branch name from a path inside a worktree.
// It walks path components under projectDir looking for a directory with a .git marker.
func DetectBranch(projectDir, cwd string) (string, error) {
	after, ok := strings.CutPrefix(cwd, projectDir+string(filepath.Separator))
	if !ok {
		return "", errors.New("path is not inside the project worktree directory")
	}
	parts := strings.Split(after, string(filepath.Separator))
	for i := 1; i <= len(parts); i++ {
		candidate := filepath.Join(projectDir, filepath.Join(parts[:i]...))
		if _, err := os.Stat(filepath.Join(candidate, ".git")); err == nil {
			return filepath.Join(parts[:i]...), nil
		}
	}
	return "", errors.New("could not detect branch from current directory")
}

// FindWorktreeGlobal searches all projects under ~/worktrees/ for a branch.
// Returns the worktree path and project name.
func FindWorktreeGlobal(branch string) (string, string, error) {
	home, _ := os.UserHomeDir()
	wtRoot := filepath.Join(home, "worktrees")
	entries, readErr := os.ReadDir(wtRoot)
	if readErr != nil {
		return "", "", fmt.Errorf("worktree not found for %q in any project", branch)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir, findErr := FindWorktree(entry.Name(), branch)
		if findErr == nil {
			return dir, entry.Name(), nil
		}
	}
	return "", "", fmt.Errorf("worktree not found for %q in any project", branch)
}

// CleanEmptyParents removes empty parent directories between path and stopAt (exclusive).
func CleanEmptyParents(path, stopAt string) {
	dir := filepath.Dir(path)
	for dir != stopAt && strings.HasPrefix(dir, stopAt) {
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) > 0 {
			break
		}
		_ = os.Remove(dir)
		dir = filepath.Dir(dir)
	}
}
