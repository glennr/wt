package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/glennr/wt/internal/repo"
	"github.com/glennr/wt/internal/worktree"

	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec <branch> [-p <project>] -- <command...>",
	Short: "Run a command in a worktree",
	Long:  "Runs a command with cwd set to the worktree directory and WT_* env vars, without changing the caller's directory.",
	Args:  cobra.MinimumNArgs(2), //nolint:mnd // branch + at least 1 command arg
	RunE:  runExec,
}

func init() {
	execCmd.Flags().StringP("project", "p", "", "Project name or source repo path")
	rootCmd.AddCommand(execCmd)
}

func runExec(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	repoFlag, _ := cmd.Flags().GetString("project")

	// args[0] is branch, args[1:] is the command (everything after --)
	branch := args[0]
	command := args[1:]

	if len(command) == 0 {
		return errors.New("usage: wt exec <branch> [-p <project>] -- <command...>")
	}

	repoPath, projectName, err := repo.Resolve(ctx, repoFlag)
	if err != nil {
		return err
	}

	wtDir, err := worktree.FindWorktree(projectName, branch)
	if err != nil {
		return fmt.Errorf("worktree for branch %q not found", branch)
	}

	c := exec.CommandContext(ctx, command[0], command[1:]...) //nolint:gosec // user-provided command is intentional
	c.Dir = wtDir
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	c.Env = append(os.Environ(),
		"WT_SOURCE_DIR="+repoPath,
		"WT_BRANCH="+branch,
		"WT_PROJECT="+projectName,
		"WT_WORKTREE_DIR="+wtDir,
	)

	err = c.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("exec failed: %w", err)
	}
	return nil
}
