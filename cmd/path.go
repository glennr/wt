package cmd

import (
	"fmt"
	"os"

	"github.com/glennr/wt/internal/repo"
	"github.com/glennr/wt/internal/worktree"

	"github.com/spf13/cobra"
)

var pathCmd = &cobra.Command{
	Use:   "path <branch>",
	Short: "Print worktree path",
	Long:  "Prints the absolute worktree path for the given branch to stdout. Exits non-zero if the worktree doesn't exist.",
	Args:  cobra.ExactArgs(1),
	RunE:  runPath,
}

func init() {
	pathCmd.Flags().StringP("project", "p", "", "Project name or source repo path")
	rootCmd.AddCommand(pathCmd)
}

func runPath(cmd *cobra.Command, args []string) error {
	projectFlag, _ := cmd.Flags().GetString("project")
	branch := args[0]

	_, projectName, err := repo.Resolve(cmd.Context(), projectFlag)
	if err != nil && projectFlag != "" {
		return err
	}

	var wtDir string
	if err == nil {
		wtDir, err = worktree.FindWorktree(projectName, branch)
	}
	// No explicit -p: fall back to searching all projects
	if err != nil && projectFlag == "" {
		wtDir, _, err = worktree.FindWorktreeGlobal(branch)
	}
	if err != nil {
		return fmt.Errorf("worktree for branch %q not found", branch)
	}

	fmt.Fprintln(os.Stdout, wtDir)
	return nil
}
