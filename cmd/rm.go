package cmd

import (
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

var rmCmd = &cobra.Command{
	Use:   "rm [<branch>] [-p <project>] [--run <cmd>] [git-flags...]",
	Short: "Remove a worktree",
	Long:  "Removes a worktree. If no argument is given and cwd is inside a worktree, removes that one. Unrecognized flags pass through to git worktree remove.",
	// We handle argument parsing manually to support git flag passthrough
	DisableFlagParsing: true,
	RunE:               runRm,
}

func init() {
	rootCmd.AddCommand(rmCmd)
}

func runRm(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	var projectFlag, runFlag string
	var gitFlags []string
	var positional []string

	for i := 0; i < len(args); i++ {
		switch {
		case (args[i] == "-p" || args[i] == "--project") && i+1 < len(args):
			projectFlag = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--project="):
			projectFlag = args[i][len("--project="):]
		case strings.HasPrefix(args[i], "-p="):
			projectFlag = args[i][len("-p="):]
		case args[i] == "--run" && i+1 < len(args):
			runFlag = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--run="):
			runFlag = args[i][len("--run="):]
		case strings.HasPrefix(args[i], "-"):
			gitFlags = append(gitFlags, args[i])
		default:
			positional = append(positional, args[i])
		}
	}

	repoPath, projectName, err := repo.Resolve(ctx, projectFlag)
	if err != nil {
		return err
	}

	var branch string
	if len(positional) > 0 {
		branch = positional[0]
	} else {
		// Auto-detect: if cwd is inside a worktree, use it
		cwd, cwdErr := os.Getwd()
		if cwdErr != nil {
			return fmt.Errorf("cannot determine current directory: %w", cwdErr)
		}
		home, _ := os.UserHomeDir()
		projectDir := filepath.Join(home, "worktrees", projectName)
		var detectErr error
		branch, detectErr = worktree.DetectBranch(projectDir, cwd)
		if detectErr != nil {
			return errors.New("no branch specified and cwd is not inside a worktree")
		}
	}

	wtDir, err := worktree.FindWorktree(projectName, branch)
	if err != nil {
		return err
	}

	// Run pre-removal command if specified
	if runFlag != "" {
		fmt.Fprintf(os.Stderr, "Running pre-removal: %s\n", runFlag)
		if err = runner.Run(ctx, runFlag, runner.RunEnv{
			SourceDir:   repoPath,
			Branch:      branch,
			Project:     projectName,
			WorktreeDir: wtDir,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "warning: %v\n", err)
		}
	}

	// Remove the worktree (git flags passed through)
	removeArgs := []string{"-C", repoPath, "worktree", "remove"}
	removeArgs = append(removeArgs, gitFlags...)
	removeArgs = append(removeArgs, wtDir)
	gitCmd := exec.CommandContext(ctx, "git", removeArgs...)
	gitCmd.Stdout = os.Stderr
	gitCmd.Stderr = os.Stderr
	if err = gitCmd.Run(); err != nil {
		return fmt.Errorf("git worktree remove failed: %w", err)
	}

	// Clean up empty intermediate directories (e.g. gr/ after removing gr/foo)
	home, _ := os.UserHomeDir()
	projectDir := filepath.Join(home, "worktrees", projectName)
	worktree.CleanEmptyParents(wtDir, projectDir)

	fmt.Fprintf(os.Stderr, "Removed worktree %s\n", wtDir)
	return nil
}
