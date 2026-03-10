package worktree

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDir(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	tests := []struct {
		project string
		branch  string
		want    string
	}{
		{"canopy", "main", filepath.Join(tmpHome, "worktrees", "canopy", "main")},
		{"api", "gr/new-endpoint", filepath.Join(tmpHome, "worktrees", "api", "gr", "new-endpoint")},
		{"api", "a/b/c", filepath.Join(tmpHome, "worktrees", "api", "a", "b", "c")},
	}
	for _, tt := range tests {
		got := Dir(tt.project, tt.branch)
		if got != tt.want {
			t.Errorf("Dir(%q, %q) = %q, want %q", tt.project, tt.branch, got, tt.want)
		}
	}
}

func TestFindWorktree(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	project := "test-project"

	// Create a worktree dir with .git marker
	branchDir := filepath.Join(tmpHome, "worktrees", project, "test-branch")
	if err := os.MkdirAll(branchDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(branchDir, ".git"), []byte("gitdir: /tmp/fake"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Should find by branch name
	got, err := FindWorktree(project, "test-branch")
	if err != nil {
		t.Fatalf("FindWorktree() error = %v", err)
	}
	if got != branchDir {
		t.Errorf("FindWorktree() = %q, want %q", got, branchDir)
	}

	// Should find by absolute path
	got, err = FindWorktree(project, branchDir)
	if err != nil {
		t.Fatalf("FindWorktree() by path error = %v", err)
	}
	if got != branchDir {
		t.Errorf("FindWorktree() by path = %q, want %q", got, branchDir)
	}

	// Should fail for nonexistent
	_, err = FindWorktree(project, "nonexistent")
	if err == nil {
		t.Error("FindWorktree() expected error for nonexistent branch")
	}
}

func TestFindWorktreeNestedBranch(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	project := "test-project"
	branchDir := filepath.Join(tmpHome, "worktrees", project, "gr", "foo")
	if err := os.MkdirAll(branchDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(branchDir, ".git"), []byte("gitdir: /tmp/fake"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got, err := FindWorktree(project, "gr/foo")
	if err != nil {
		t.Fatalf("FindWorktree() error = %v", err)
	}
	if got != branchDir {
		t.Errorf("FindWorktree() = %q, want %q", got, branchDir)
	}
}

func TestFindWorktreeBrokenMetadata(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	project := "test-project"
	branchDir := filepath.Join(tmpHome, "worktrees", project, "broken-branch")
	if err := os.MkdirAll(branchDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	// No .git marker

	_, err := FindWorktree(project, "broken-branch")
	if err == nil {
		t.Error("FindWorktree() expected error for broken worktree")
	}
}

func TestListWorktrees(t *testing.T) {
	projectDir := t.TempDir()

	// Create a simple worktree
	wt1 := filepath.Join(projectDir, "main")
	os.MkdirAll(wt1, 0o755)
	os.WriteFile(filepath.Join(wt1, ".git"), []byte("gitdir: /tmp/fake"), 0o644)

	// Create a nested worktree
	wt2 := filepath.Join(projectDir, "gr", "foo")
	os.MkdirAll(wt2, 0o755)
	os.WriteFile(filepath.Join(wt2, ".git"), []byte("gitdir: /tmp/fake"), 0o644)

	// Create a plain dir (not a worktree)
	os.MkdirAll(filepath.Join(projectDir, "not-a-worktree"), 0o755)

	entries := ListWorktrees(projectDir)
	if len(entries) != 2 {
		t.Fatalf("ListWorktrees() returned %d entries, want 2", len(entries))
	}

	branches := map[string]bool{}
	for _, e := range entries {
		branches[e.Branch] = true
	}
	if !branches["main"] {
		t.Error("ListWorktrees() missing 'main'")
	}
	if !branches[filepath.Join("gr", "foo")] {
		t.Errorf("ListWorktrees() missing 'gr/foo', got %v", branches)
	}
}

func TestDetectBranch(t *testing.T) {
	projectDir := t.TempDir()

	// Create nested worktree
	wtDir := filepath.Join(projectDir, "gr", "foo")
	os.MkdirAll(filepath.Join(wtDir, "src"), 0o755)
	os.WriteFile(filepath.Join(wtDir, ".git"), []byte("gitdir: /tmp/fake"), 0o644)

	// Detect from inside worktree subdirectory
	branch, err := DetectBranch(projectDir, filepath.Join(wtDir, "src"))
	if err != nil {
		t.Fatalf("DetectBranch() error = %v", err)
	}
	if branch != filepath.Join("gr", "foo") {
		t.Errorf("DetectBranch() = %q, want %q", branch, filepath.Join("gr", "foo"))
	}

	// Detect from worktree root
	branch, err = DetectBranch(projectDir, wtDir)
	if err != nil {
		t.Fatalf("DetectBranch() from root error = %v", err)
	}
	if branch != filepath.Join("gr", "foo") {
		t.Errorf("DetectBranch() from root = %q, want %q", branch, filepath.Join("gr", "foo"))
	}
}

func TestDetectBranchSimple(t *testing.T) {
	projectDir := t.TempDir()

	wtDir := filepath.Join(projectDir, "main")
	os.MkdirAll(filepath.Join(wtDir, "src"), 0o755)
	os.WriteFile(filepath.Join(wtDir, ".git"), []byte("gitdir: /tmp/fake"), 0o644)

	branch, err := DetectBranch(projectDir, filepath.Join(wtDir, "src"))
	if err != nil {
		t.Fatalf("DetectBranch() error = %v", err)
	}
	if branch != "main" {
		t.Errorf("DetectBranch() = %q, want %q", branch, "main")
	}
}

func TestFindWorktreeGlobal(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create worktrees in two projects
	wt1 := filepath.Join(tmpHome, "worktrees", "canopy", "gr", "foo")
	os.MkdirAll(wt1, 0o755)
	os.WriteFile(filepath.Join(wt1, ".git"), []byte("gitdir: /tmp/fake"), 0o644)

	wt2 := filepath.Join(tmpHome, "worktrees", "api", "main")
	os.MkdirAll(wt2, 0o755)
	os.WriteFile(filepath.Join(wt2, ".git"), []byte("gitdir: /tmp/fake"), 0o644)

	// Should find gr/foo in canopy
	dir, project, err := FindWorktreeGlobal("gr/foo")
	if err != nil {
		t.Fatalf("FindWorktreeGlobal(gr/foo) error = %v", err)
	}
	if dir != wt1 {
		t.Errorf("FindWorktreeGlobal(gr/foo) dir = %q, want %q", dir, wt1)
	}
	if project != "canopy" {
		t.Errorf("FindWorktreeGlobal(gr/foo) project = %q, want %q", project, "canopy")
	}

	// Should find main in api
	dir, project, err = FindWorktreeGlobal("main")
	if err != nil {
		t.Fatalf("FindWorktreeGlobal(main) error = %v", err)
	}
	if dir != wt2 {
		t.Errorf("FindWorktreeGlobal(main) dir = %q, want %q", dir, wt2)
	}
	if project != "api" {
		t.Errorf("FindWorktreeGlobal(main) project = %q, want %q", project, "api")
	}

	// Should fail for nonexistent
	_, _, err = FindWorktreeGlobal("nonexistent")
	if err == nil {
		t.Error("FindWorktreeGlobal(nonexistent) expected error")
	}
}

func TestFindWorktreeGlobalNoWorktreesDir(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	_, _, err := FindWorktreeGlobal("anything")
	if err == nil {
		t.Error("FindWorktreeGlobal() expected error when ~/worktrees/ doesn't exist")
	}
}

func TestCleanEmptyParents(t *testing.T) {
	tmpDir := t.TempDir()
	nested := filepath.Join(tmpDir, "a", "b", "c")
	os.MkdirAll(nested, 0o755)
	// Simulate git worktree remove having deleted the leaf
	os.Remove(nested)

	CleanEmptyParents(nested, tmpDir)

	if _, err := os.Stat(filepath.Join(tmpDir, "a")); !os.IsNotExist(err) {
		t.Error("expected 'a' to be removed")
	}
}

func TestCleanEmptyParentsStopsAtNonEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "a", "b", "c"), 0o755)
	os.MkdirAll(filepath.Join(tmpDir, "a", "other"), 0o755)
	// Simulate git worktree remove having deleted the leaf
	os.Remove(filepath.Join(tmpDir, "a", "b", "c"))

	CleanEmptyParents(filepath.Join(tmpDir, "a", "b", "c"), tmpDir)

	if _, err := os.Stat(filepath.Join(tmpDir, "a", "b")); !os.IsNotExist(err) {
		t.Error("expected 'b' to be removed")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "a")); os.IsNotExist(err) {
		t.Error("expected 'a' to remain (has other sibling)")
	}
}
