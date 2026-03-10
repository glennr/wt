package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/glennr/wt/internal/worktree"

	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List worktrees",
	Long:  "Lists worktrees. With -p, lists only that project's worktrees. Without, scans ~/worktrees/ for all projects.",
	RunE:  runLs,
}

func init() {
	lsCmd.Flags().StringP("project", "p", "", "Filter to one project")
	rootCmd.AddCommand(lsCmd)
}

func runLs(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	projectFlag, _ := cmd.Flags().GetString("project")

	home, _ := os.UserHomeDir()
	wtRoot := filepath.Join(home, "worktrees")

	const tabWidth = 4
	const padding = 2
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, tabWidth, padding, ' ', 0)
	fmt.Fprintln(w, "PROJECT\tBRANCH\tCOMMIT\tMODIFIED\tPATH")

	if projectFlag != "" {
		project := resolveProjectName(projectFlag)
		projectDir := filepath.Join(wtRoot, project)
		printProjectWorktrees(ctx, w, project, projectDir)
		return w.Flush()
	}

	entries, err := os.ReadDir(wtRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return w.Flush()
		}
		return fmt.Errorf("reading %s: %w", wtRoot, err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			projectDir := filepath.Join(wtRoot, entry.Name())
			printProjectWorktrees(ctx, w, entry.Name(), projectDir)
		}
	}

	return w.Flush()
}

// resolveProjectName extracts the project name from a -p flag value.
func resolveProjectName(flag string) string {
	if strings.ContainsAny(flag, "/.~") {
		if strings.HasPrefix(flag, "~/") {
			home, _ := os.UserHomeDir()
			flag = filepath.Join(home, flag[2:])
		}
		return filepath.Base(flag)
	}
	return flag
}

func printProjectWorktrees(ctx context.Context, w *tabwriter.Writer, project, projectDir string) {
	entries := worktree.ListWorktrees(projectDir)
	for _, entry := range entries {
		commit := getShortCommit(ctx, entry.Path)
		modified := getLastModified(ctx, entry.Path)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", project, entry.Branch, commit, modified, entry.Path)
	}
}

func getShortCommit(ctx context.Context, wtPath string) string {
	out, err := exec.CommandContext(ctx, "git", "-C", wtPath, "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return "???????"
	}
	return strings.TrimSpace(string(out))
}

func getLastModified(ctx context.Context, wtPath string) string {
	out, err := exec.CommandContext(ctx, "git", "-C", wtPath, "log", "-1", "--format=%ct").Output()
	if err != nil {
		return "unknown"
	}
	epoch, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return "unknown"
	}
	return relativeTime(time.Unix(epoch, 0))
}

const (
	hoursPerDay  = 24
	daysPerMonth = 30
)

func relativeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		n := int(d.Minutes())
		if n == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", n)
	case d < 24*time.Hour:
		n := int(d.Hours())
		if n == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", n)
	case d < daysPerMonth*hoursPerDay*time.Hour:
		n := int(d.Hours() / hoursPerDay)
		if n == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", n)
	default:
		n := int(d.Hours() / hoursPerDay / daysPerMonth)
		if n == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", n)
	}
}
