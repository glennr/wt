package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

// Version is set via ldflags at build time.
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:     "wt",
	Short:   "A general-purpose git worktree manager",
	Long:    "wt wraps git worktree commands, adds central worktree storage (~~/worktrees/<project>/<branch>), and supports project-specific bootstrap/teardown via --run.",
	Version: Version,
}

func Execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}
