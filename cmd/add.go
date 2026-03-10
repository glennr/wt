package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/glennr/wt/internal/repo"
	"github.com/glennr/wt/internal/runner"
	"github.com/glennr/wt/internal/worktree"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [flags...] <branch>",
	Short: "Create or return a worktree",
	Long:  "Idempotent create-or-return. Always prints the worktree path. Creates branch+worktree if needed.",
	// We handle argument parsing manually to support git flag passthrough
	DisableFlagParsing: true,
	RunE:               runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)
}

type addFlags struct {
	project  string
	run      string
	branch   string
	gitArgs  []string
	hasBFlag bool
}

func parseAddFlags(args []string) addFlags {
	var f addFlags
	for i := 0; i < len(args); i++ {
		switch {
		case (args[i] == "-p" || args[i] == "--project") && i+1 < len(args):
			f.project = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--project="):
			f.project = args[i][len("--project="):]
		case strings.HasPrefix(args[i], "-p="):
			f.project = args[i][len("-p="):]
		case args[i] == "--run" && i+1 < len(args):
			f.run = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--run="):
			f.run = args[i][len("--run="):]
		case args[i] == "-b" && i+1 < len(args):
			f.hasBFlag = true
			f.branch = args[i+1]
			f.gitArgs = append(f.gitArgs, args[i], args[i+1])
			i++
		default:
			f.gitArgs = append(f.gitArgs, args[i])
		}
	}
	if f.branch == "" && len(f.gitArgs) > 0 {
		f.branch = f.gitArgs[len(f.gitArgs)-1]
	}
	return f
}

func runAdd(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	f := parseAddFlags(args)

	if f.branch == "" {
		return errors.New("branch name is required")
	}

	repoPath, projectName, err := repo.Resolve(ctx, f.project)
	if err != nil {
		return err
	}

	wtDir := worktree.Dir(projectName, f.branch)

	// Idempotent: if worktree already exists with valid metadata, return its path
	if _, statErr := os.Stat(filepath.Join(wtDir, ".git")); statErr == nil {
		fmt.Fprintln(os.Stdout, wtDir)
		return nil
	}
	// Directory exists but no .git marker = broken metadata
	if _, statErr := os.Stat(wtDir); statErr == nil {
		return fmt.Errorf("directory %s exists but is not a valid worktree (missing .git metadata)", wtDir)
	}

	gitCmdArgs := buildWorktreeAddArgs(ctx, repoPath, wtDir, f)

	gitCmd := exec.CommandContext(ctx, "git", gitCmdArgs...)
	gitCmd.Stderr = os.Stderr
	if err = gitCmd.Run(); err != nil {
		return fmt.Errorf("git worktree add failed: %w", err)
	}

	// Run post-creation command if specified
	if f.run != "" {
		fmt.Fprintf(os.Stderr, "Running: %s\n", f.run)
		if err = runner.Run(ctx, f.run, runner.RunEnv{
			SourceDir:   repoPath,
			Branch:      f.branch,
			Project:     projectName,
			WorktreeDir: wtDir,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "warning: %v\n", err)
		}
	}

	fmt.Fprintln(os.Stdout, wtDir)
	return nil
}

func buildWorktreeAddArgs(ctx context.Context, repoPath, wtDir string, f addFlags) []string {
	if f.hasBFlag {
		return append([]string{"-C", repoPath, "worktree", "add", wtDir}, f.gitArgs...)
	}
	var passthroughFlags []string
	if len(f.gitArgs) > 1 {
		passthroughFlags = f.gitArgs[:len(f.gitArgs)-1]
	}
	base := append([]string{"-C", repoPath, "worktree", "add"}, passthroughFlags...)
	if branchExistsLocallyOrRemote(ctx, repoPath, f.branch) {
		return append(base, wtDir, f.branch)
	}
	return append(base, "-b", f.branch, wtDir)
}

func branchExistsLocallyOrRemote(ctx context.Context, repoPath, branch string) bool {
	// Check local branches
	if exec.CommandContext(ctx, "git", "-C", repoPath, "show-ref", "--verify", "--quiet", "refs/heads/"+branch).
		Run() ==
		nil {
		return true
	}
	// Check remote branches
	out, err := exec.CommandContext(ctx, "git", "-C", repoPath, "branch", "-r", "--list", "*/"+branch).Output()
	return err == nil && len(strings.TrimSpace(string(out))) > 0
}
