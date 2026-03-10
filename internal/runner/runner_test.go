package runner

import (
	"context"
	"testing"
)

func TestRun(t *testing.T) {
	tmpDir := t.TempDir()

	err := Run(context.Background(), "true", RunEnv{
		SourceDir:   "/tmp/source",
		Branch:      "test-branch",
		Project:     "test-project",
		WorktreeDir: tmpDir,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestRunFailure(t *testing.T) {
	tmpDir := t.TempDir()

	err := Run(context.Background(), "false", RunEnv{
		SourceDir:   "/tmp/source",
		Branch:      "test-branch",
		Project:     "test-project",
		WorktreeDir: tmpDir,
	})
	if err == nil {
		t.Error("Run() expected error for failing command")
	}
}

func TestRunEnvVars(t *testing.T) {
	tmpDir := t.TempDir()

	// Verify env vars are set by echoing them
	cmd := `test "$WT_SOURCE_DIR" = "/tmp/source" && ` +
		`test "$WT_BRANCH" = "test-branch" && ` +
		`test "$WT_PROJECT" = "test-project"`
	err := Run(context.Background(), cmd, RunEnv{
		SourceDir:   "/tmp/source",
		Branch:      "test-branch",
		Project:     "test-project",
		WorktreeDir: tmpDir,
	})
	if err != nil {
		t.Fatalf("Run() env vars not set correctly: %v", err)
	}
}
