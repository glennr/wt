package cmd

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var pruneCmd = &cobra.Command{
	Use:   "prune [flags]",
	Short: "Prune stale worktree references",
	Long:  "Passthrough to git worktree prune.",
	RunE:  runPrune,
}

func init() {
	pruneCmd.DisableFlagParsing = true
	rootCmd.AddCommand(pruneCmd)
}

func runPrune(cmd *cobra.Command, args []string) error {
	gitArgs := append([]string{"worktree", "prune"}, args...)
	gitCmd := exec.CommandContext(cmd.Context(), "git", gitArgs...)
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr
	return gitCmd.Run()
}
