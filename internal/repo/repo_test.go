package repo

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestResolveWithPath(t *testing.T) {
	tmpDir := t.TempDir()

	repoPath, projectName, err := Resolve(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	abs, _ := filepath.Abs(tmpDir)
	if repoPath != abs {
		t.Errorf("repoPath = %q, want %q", repoPath, abs)
	}
	if projectName != filepath.Base(abs) {
		t.Errorf("projectName = %q, want %q", projectName, filepath.Base(abs))
	}
}

func TestResolveFromGitRepo(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize a git repo
	cmd := exec.CommandContext(context.Background(), "git", "init", tmpDir)
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init failed: %v", err)
	}

	// Change to the repo directory
	t.Chdir(tmpDir)

	repoPath, projectName, err := Resolve(context.Background(), "")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if repoPath != tmpDir {
		t.Errorf("repoPath = %q, want %q", repoPath, tmpDir)
	}
	if projectName != filepath.Base(tmpDir) {
		t.Errorf("projectName = %q, want %q", projectName, filepath.Base(tmpDir))
	}
}

func TestResolveNoRepo(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	_, _, err := Resolve(context.Background(), "")
	if err == nil {
		t.Error("Resolve() expected error when not in a git repo")
	}
}

func TestIsPath(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"canopy", false},
		{"my-project", false},
		{"~/src/canopy", true},
		{"/home/user/src/canopy", true},
		{"./canopy", true},
		{"../canopy", true},
	}
	for _, tt := range tests {
		got := isPath(tt.input)
		if got != tt.want {
			t.Errorf("isPath(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestExpandTilde(t *testing.T) {
	home, _ := os.UserHomeDir()
	tests := []struct {
		input string
		want  string
	}{
		{"~/src/canopy", filepath.Join(home, "src", "canopy")},
		{"/absolute/path", "/absolute/path"},
		{"relative", "relative"},
	}
	for _, tt := range tests {
		got := expandTilde(tt.input)
		if got != tt.want {
			t.Errorf("expandTilde(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
